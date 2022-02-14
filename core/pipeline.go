package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sharvanath/kromium/storage"
	"github.com/sharvanath/kromium/transforms"
	log "github.com/sirupsen/logrus"
	"io"
)

// Returns the number of files copied.
func RunPipeline(ctx context.Context, config *PipelineConfig) (int, error) {
	inputStorageProvider := storage.GetStorageProvider(config.SourceBucket)
	outputStorageProvider := storage.GetStorageProvider(config.DestinationBucket)
	copied := 0
	if inputStorageProvider == nil {
		return copied, fmt.Errorf("Input storage provide not found for %s.", config.SourceBucket)
	}
	files, err := inputStorageProvider.ListObjects(ctx, config.SourceBucket)
	if err != nil {
		return copied, err
	}

	if len(config.Transforms) == 0 {
		return copied, fmt.Errorf("NOOP: Empty pipeline")
	}

	workerState, err := ReadMergedState(ctx, config, len(files))
	if err != nil {
		return copied, err
	}

	start, end := workerState.findProcessingRange()
	if start == -1 {
		log.Infof("All files have been processed. %d\n", len(files))
		return copied, nil
	}

	workerId := uuid.New().String()
	log.Debugf("Starting worker %s with index range %d:%d\n", workerId, start, end)
	for _, o := range files[start:end] {
		log.Debugf("Processing object: %s from bucket: %s\n", o, config.SourceBucket)
		var buf *bytes.Buffer
		for idx, t := range config.Transforms {
			log.Debugf("Apply transform [%2d] %15s.", idx, t)
			if idx != 0 {
				log.Debugf(" Input for stage [%2d]: %d.", idx, buf.Len())
			}
			var src io.Reader
			var dst io.Writer

			if idx == 0 {
				srcCloser, err := inputStorageProvider.ObjectReader(ctx, config.SourceBucket, o)
				if err != nil {
					return copied, err
				}
				src = srcCloser
				defer srcCloser.Close()
			} else {
				src = bufio.NewReader(buf)
			}

			if idx == len(config.Transforms)-1 {
				dstCloser, err := outputStorageProvider.ObjectWriter(ctx, config.DestinationBucket, o+config.NameSuffix)
				if err != nil {
					return copied, err
				}
				defer dstCloser.Close()
				dst = dstCloser
			} else {
				var output bytes.Buffer
				dst = &output
				buf = &output
			}

			transform := transforms.GetTransform(t.Name, t.Args)
			if transform == nil {
				return copied, fmt.Errorf("Could not find transform %s.", t.Name)
			}
			if _, err := transform.Transform(dst, src); err != nil {
				return copied, err
			}

			if idx != len(config.Transforms)-1 {
				log.Debugf(" Output for stage [%2d]: %d\n", idx, buf.Len())
			} else {
				log.Debugf("\n")
			}
		}
		copied += 1
		log.Debugf("Wrote object: %s to bucket: %s\n", o+config.NameSuffix, config.DestinationBucket)
	}

	workerState.setProcessed(start)
	workerState.workerId = workerId
	return copied, WriteState(ctx, config.StateBucket, workerState)
}

func runPipelineLoopInternal(ctx context.Context, config *PipelineConfig, channel chan error) {
	total := 0
	for {
		count, err := RunPipeline(ctx, config)
		if err != nil {
			channel <- err
			return
		}
		total += count
		if count <= 0 {
			break
		}
	}
	channel <- nil
}

func RunPipelineLoopParallel(ctx context.Context, config *PipelineConfig, parallelism int) error {
	var channels []chan error
	for i := 0; i < parallelism; i += 1 {
		channel := make(chan error)
		channels = append(channels, channel)
		go runPipelineLoopInternal(ctx, config, channel)
	}

	for _, channel := range channels {
		e := <-channel
		if e != nil {
			return e
		}
	}

	return nil
}

func RunPipelineLoop(ctx context.Context, config *PipelineConfig) error {
	return RunPipelineLoopParallel(ctx, config, 1)
}
