package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/uuid"
	"github.com/sharvanath/kromium/transforms"
	log "github.com/sirupsen/logrus"
	"io"
	//"runtime/trace"
	"sync/atomic"
	"time"
)

var processedCount int32

func updateStatus(text string, renderUI bool) {
	if !renderUI {
		fmt.Print(text + "\n")
		return
	}
	topBox := widgets.NewParagraph()
	topBox.Text = text
	topBox.TextStyle.Fg = 0
	topBox.SetRect(0, 0, 45,  3)
	ui.Render(topBox)
}

func updateWorkerStatus(idx int, text string, renderUI bool) {
	if !renderUI {
		fmt.Printf("worker %d: %s\n", idx, text)
		return
	}
	topBox := widgets.NewParagraph()
	topBox.Text = text
	topBox.TextStyle.Fg = 0
	topBox.SetRect(0, idx * 3 + 3, 45,  idx * 3 + 6)
	ui.Render(topBox)
}


// Returns the number of files copied.
func RunPipeline(ctx context.Context, config *PipelineConfig, threadIdx int, renderUi bool) (int, error) {
	copied := 0
	files, err := config.sourceStorageProvider.ListObjects(ctx, config.SourceBucket)
	if err != nil {
		return copied, err
	}

	if len(config.Transforms) == 0 || len(files) == 0  {
		return copied, fmt.Errorf("NOOP: Empty pipeline")
	}

	workerState, err := ReadMergedState(ctx, config, len(files))
	//r.End()

	if err != nil {
		return copied, err
	}

	start, end := workerState.findProcessingRange()
	if start == -1 {
		log.Debugf("[Worker %d] All files have been processed. %d\n", threadIdx, len(files))
		return copied, nil
	}

	workerId := uuid.New().String()
	log.Debugf("[Worker %d] Starting worker %s with index range %d:%d\n", threadIdx, workerId, start, end)
	var channels []chan error

	for _, o1 := range files[start:end] {
		//trace.Start(os.Stderr)
		log.Debugf("[Worker %d] Processing object: %s from bucket: %s\n", threadIdx, o1, config.SourceBucket)
		var buf *bytes.Buffer
		var err error
		var srcCloser io.ReadCloser
		var dstCloser io.WriteCloser

		channel := make(chan error)
		channels = append(channels, channel)
		go func(o string, c chan error) {
			for idx, t := range config.Transforms {
				//r := trace.StartRegion(ctx, t.Type)
				log.Debugf("[Worker %d] Apply transform [%2d] %15s.", threadIdx, idx, t)
				if idx != 0 {
					log.Debugf("[Worker %d] Input for stage [%2d]: %d.", threadIdx, idx, buf.Len())
				}
				var src io.Reader
				var dst io.Writer

				if idx == 0 {
					srcCloser, err = config.sourceStorageProvider.ObjectReader(ctx, config.SourceBucket, o)
					if err != nil {
						srcCloser = nil
						break
					}
					src = srcCloser
				} else {
					src = bufio.NewReader(buf)
				}

				if idx == len(config.Transforms)-1 {
					dstCloser, err = config.destStorageProvider.ObjectWriter(ctx, config.DestinationBucket, o+config.NameSuffix)
					if err != nil {
						dstCloser = nil
						break
					}
					dst = dstCloser
				} else {
					var output bytes.Buffer
					dst = &output
					buf = &output
				}

				transform := transforms.GetTransform(t.Type, t.Args)
				if transform == nil {
					err = fmt.Errorf("Could not find transform %s.", t.Type)
					break
				}

				if _, err := transform.Transform(dst, src); err != nil {
					break
				}

				if idx != len(config.Transforms)-1 {
					log.Debugf("[Worker %d] Output for stage [%2d]: %d\n", threadIdx, idx, buf.Len())
				} else {
					log.Debugf("\n")
				}

				srcCloser.Close()
				dstCloser.Close()
			}

			if err != nil {
				log.Warnf("[Worker %d] Failed during pipeline %s", threadIdx, err)
				if srcCloser != nil {
					srcCloser.Close()
				}
				if dstCloser != nil {
					dstCloser.Close()
				}
			}
			c <- err
			log.Debugf("[Worker %d] Wrote object: %s to bucket: %s\n", threadIdx, o+config.NameSuffix, config.DestinationBucket)
		}(o1, channel)
	}

	for _, c := range channels {
		e := <- c
		if e != nil {
			return copied, e
		}
		copied += 1
	}

	workerState.setProcessed(start)
	workerState.workerId = workerId
	updateStatus(fmt.Sprintf("[%s] [%d] Done %d/%d", time.Now().UTC().String(), threadIdx, workerState.m.usedSize() * cBatchSize, len(workerState.m.slice) * cBatchSize * 8), renderUi)
	return copied, WriteState(ctx, config.StateBucket, workerState)
}

func runPipelineLoopInternal(ctx context.Context, config *PipelineConfig, channel chan error, threadIdx int, renderUi bool) {
	total := 0
	for {
		count, err := RunPipeline(ctx, config, threadIdx, renderUi)
		if err != nil {
			channel <- err
		}
		if count <= 0 {
			break
		}
		total += count
		//updateWorkerStatus(threadIdx, fmt.Sprintf("[Worker %d] Processed %d objects.", threadIdx, total), renderUi)
	}
	atomic.AddInt32(&processedCount, int32(total))
	channel <- nil
}

func RunPipelineLoop(ctx context.Context, config *PipelineConfig, parallelism int, renderUi bool) error {
	if renderUi {
		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}
	}

	if parallelism <= 0 {
		return fmt.Errorf("illegal parallelism: %d", parallelism)
	}

	start := time.Now()
	var channels []chan error
	for i := 0; i < parallelism; i += 1 {
		channel := make(chan error)
		channels = append(channels, channel)
		go runPipelineLoopInternal(ctx, config, channel, i, renderUi)
	}

	for _, channel := range channels {
		e := <-channel
		if e != nil {
			if renderUi {
				ui.Close()
			}
			return e
		}
	}

	if renderUi {
		ui.Close()
	}
	fmt.Printf("Processed %d files in %.2f seconds\n", processedCount, time.Since(start).Seconds())
	return nil
}
