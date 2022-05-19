package core

import (
	"context"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/uuid"
	"github.com/sharvanath/kromium/storage"
	"github.com/sharvanath/kromium/transforms"
	log "github.com/sirupsen/logrus"
	"io"
	"runtime/trace"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var processedCount int32
var lastPct int32

func updateStatus(percent int, message string, renderUI bool) {
	if !renderUI {
		return
	}

	if int32(percent) < atomic.LoadInt32(&lastPct) {
		// other workers could see smaller values, avoid unnecessary fluctuation in the progress bar.
		return
	}

	atomic.StoreInt32(&lastPct, int32(percent))


	p := widgets.NewParagraph()
	p.Text = message
	p.TextStyle.Fg = ui.ColorBlack
	p.SetRect(0, 3, 100, 6)
	ui.Render(p)

	g0 := widgets.NewGauge()
	g0.Title = "Progress"
	g0.SetRect(0, 6, 100, 11)
	g0.Percent = percent
	g0.BarColor = ui.ColorRed
	g0.BorderStyle.Fg = ui.ColorWhite
	g0.TitleStyle.Fg = ui.ColorCyan
	ui.Render(g0)
}

func updateWorkerStatus(idx int, text string) {
	/*topBox := widgets.NewParagraph()
	topBox.Text = text
	topBox.TextStyle.Fg = 0
	topBox.SetRect(0, idx * 3 + 3, 45,  idx * 3 + 6)
	ui.Render(topBox)*/
}

func getObjectName(object string, nameSuffix string, stripSuffix string) string {
	return strings.TrimSuffix(object, stripSuffix) + nameSuffix
}

func processObjectInPipeline(ctx context.Context, config *PipelineConfig, threadIdx int, object string) error {
	srcObjectCloser, err := storage.GetObjectReader(ctx, config.sourceStorageProvider, config.SourceBucket, object)
	if err != nil {
		return err
	}
	dstObjectName := getObjectName(object, config.NameSuffix, config.StripSuffix)
	dstObjectCloser, err := storage.GetObjectWriter(ctx, config.destStorageProvider, config.DestinationBucket, dstObjectName)
	if err != nil {
		srcObjectCloser.Close()
		return err
	}

	var lastPipeReadEnd io.Reader
	// A pipeline of transforms, chained. Each stage is connected by a pipe, so the writer end must close otherwise
	// the read will keep hanging.
	var pipelineError atomic.Value
	var wg sync.WaitGroup

	for idx, t := range config.Transforms {
		var src io.Reader
		var dst io.WriteCloser
		transform := transforms.GetTransform(t.Type, t.Args)
		// Stop building the pipeline. srcObjectCloser.Close() will trigger the close of the full chain.
		if transform == nil {
			err = fmt.Errorf("could not find transform %s", t.Type)
			dstObjectCloser.Close()
			break
		}

		if idx == 0 {
			src = srcObjectCloser
		} else {
			src = lastPipeReadEnd
		}

		if idx == len(config.Transforms)-1 {
			dst = dstObjectCloser
		} else {
			lastPipeReadEnd, dst = io.Pipe()
		}

		wg.Add(1)
		go func(dst io.WriteCloser, src io.Reader, t TransformConfig) {
			log.Debugf("[Worker %d] Apply transform [%2d] %15s.", threadIdx, idx, t)
			if _, localErr := transform.Transform(dst, src);  localErr != nil {
				pipelineError.Store(localErr)
				log.Warnf("[Worker %d] Apply transform [%2d] %15s failed on %s.", threadIdx, idx, t, object)
			}
			dst.Close()
			wg.Done()
		}(dst, src, t)
	}

	wg.Wait()
	if pipelineError.Load() != nil {
		err = pipelineError.Load().(error)
		log.Warnf("[Worker %d] Failed during pipeline %s", threadIdx, err)
	}
	log.Debugf("[Worker %d] Wrote object: %s to bucket: %s\n", threadIdx, dstObjectName, config.DestinationBucket)
	return err
}

// Returns the number of files copied, and error if it fails.
func RunPipeline(ctx context.Context, config *PipelineConfig, threadIdx int, renderUi bool) (int, error) {
	defer trace.StartRegion(ctx, "RunPipeline").End()

	copied := 0
	files, err := storage.ListObjects(ctx, config.sourceStorageProvider, config.SourceBucket)
	if err != nil {
		return copied, err
	}

	if len(config.Transforms) == 0 || len(files) == 0  {
		return copied, fmt.Errorf("NOOP: Empty pipeline")
	}

	workerState, err := ReadMergedState(ctx, config, len(files))

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
		log.Debugf("[Worker %d] Processing object: %s from bucket: %s\n", threadIdx, o1, config.SourceBucket)
		var err error
		var srcCloser io.ReadCloser
		var dstCloser io.WriteCloser

		channel := make(chan error)
		channels = append(channels, channel)
		go func(o string, c chan error) {
			err = processObjectInPipeline(ctx, config, threadIdx, o)
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

	numProcessed := workerState.m.usedSize()*cBatchSize
	numTotal := len(workerState.m.slice)*cBatchSize*8
	if !renderUi {
		log.Infof("[%s] [%d] Done %d/%d", time.Now().Format("2006-01-02 15:04:05.00"), threadIdx, numProcessed, numTotal)
	}
	updateStatus((numProcessed*100)/numTotal, fmt.Sprintf("Processed %d/%d", numProcessed, numTotal), renderUi)
	fmt.Printf("%d", copied)
	return copied, WriteState(ctx, config.StateBucket, workerState)
}

func runPipelineLoopInternal(ctx context.Context, config *PipelineConfig, channel chan error, threadIdx int, renderUi bool) {
	total := 0
	for {
		count, err := RunPipeline(ctx, config, threadIdx, renderUi)
		if err != nil {
			channel <- err
			break
		}
		if count <= 0 {
			break
		}
		total += count
		if renderUi {
			updateWorkerStatus(threadIdx, fmt.Sprintf("[%s] [%d] Processed %d objects.", time.Now().Format("2006-01-02 15:04:05.00"), threadIdx, total))
		}
	}
	atomic.AddInt32(&processedCount, int32(total))
	channel <- nil
}

func RunPipelineLoop(ctx context.Context, config *PipelineConfig, parallelism int, renderUi bool) error {

	if renderUi {
		if err := ui.Init(); err != nil {
			log.Fatalf("failed to initialize termui: %v", err)
		}

		p := widgets.NewParagraph()
		p.Title = "Kromium run (https://kromium.io)"
		p.Text = fmt.Sprintf("Press any key to exit. Simply restart to resume. Using %d workers", parallelism)
		p.TextStyle.Fg = ui.ColorBlack
		p.SetRect(0, 0, 100, 3)
		ui.Render(p)
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
			return e
		}
	}

	if renderUi {
		ui.Close()
	}
	fmt.Printf("Processed %d files in %.2f seconds\n", processedCount, time.Since(start).Seconds())
	return nil
}