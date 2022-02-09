package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sharvanath/kromium/storage"
	"github.com/sharvanath/kromium/transforms"
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

	idx := workerState.findRandomEmpty()
	if idx == -1 {
		fmt.Printf("All files have been processed. %d\n",len(files))
		return copied, nil
	}

	end := idx + 8
	if end > len(files) {
		end = len(files)
	}

	workerId := uuid.New().String()

	fmt.Printf("Starting worker %s with index range %d:%d\n", workerId, idx, end)
	for _, o := range files[idx:end] {
		fmt.Printf("Processing object: %s from bucket: %s\n", o, config.SourceBucket)
		var buf *bytes.Buffer
		for idx, t := range config.Transforms {
			fmt.Printf("Apply transform [%2d] %15s.", idx, t)
			if idx != 0 {
				fmt.Printf(" Input for stage [%2d]: %d.", idx, buf.Len())
			}
			var src io.Reader
			var dst io.Writer

			if idx == 0 {
				srcCloser, err := inputStorageProvider.ObjectReader(ctx, config.SourceBucket, o)
				if err != nil {
					return  copied, err
				}
				src = srcCloser
				defer srcCloser.Close()
			} else {
				src = bufio.NewReader(buf)
			}

			if idx == len(config.Transforms) - 1 {
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

			if idx != len(config.Transforms) - 1 {
				fmt.Printf(" Output for stage [%2d]: %d\n", idx, buf.Len())
			} else {
				fmt.Printf("\n")
			}
		}
		copied += 1
		fmt.Printf("Wrote object: %s to bucket: %s\n", o+config.NameSuffix, config.DestinationBucket)
	}

	workerState.bitmap[idx >> 3] = 1
	workerState.workerId = workerId
	return copied, WriteState(ctx, config.StateBucket, workerState)
}

func RunPipelineLoop(ctx context.Context, config *PipelineConfig) error {
	for ; ; {
		count, err := RunPipeline(ctx, config)
		if err != nil {
			return err
		}
		if count <= 0 {
			break
		}
	}
	return nil
}