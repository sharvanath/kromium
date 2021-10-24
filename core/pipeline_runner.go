package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/sharvanath/kromium/storage"
	"github.com/sharvanath/kromium/transforms"
	"io"
)



func RunPipeline(ctx context.Context, config PipelineConfig) error {
	inputStorageProvider := storage.GetStorageProvider(config.SourceBucket)
	outputStorageProvider := storage.GetStorageProvider(config.DestinationBucket)
	files, err := inputStorageProvider.ListObjects(ctx, config.SourceBucket)
	if err != nil {
		return err
	}

	if len(config.Transforms) == 0 {
		return fmt.Errorf("NOOP: Empty pipeline")
	}

	for _, o := range files {
		fmt.Printf("Processing object: %s from bucket: %s\n", o, config.SourceBucket)
		var buf *bytes.Buffer
		for idx, t := range config.Transforms {
			fmt.Printf("Apply transform [%d] %s\n", idx, t)
			var src io.Reader
			var dst io.Writer

			if idx == 0 {
				srcCloser, err := inputStorageProvider.ObjectReader(ctx, config.SourceBucket, o)
				defer srcCloser.Close()
				src = srcCloser
				if err != nil {
					return  err
				}
			} else {
				src = bufio.NewReader(buf)
			}

			if idx == len(config.Transforms) - 1 {
				dstCloser, err := outputStorageProvider.ObjectWriter(ctx, config.DestinationBucket, o+config.NameSuffix)
				defer dstCloser.Close()
				dst = dstCloser
				if err != nil {
					return err
				}
			} else {
				var output bytes.Buffer
				dst = bufio.NewWriter(&output)
				buf = &output
			}

			transform := transforms.GetTransform(t.Name, t.Args)
			if transform == nil {
				return fmt.Errorf("Could not find transform %s.", t.Name)
			}
			if _, err := transform.Transform(dst, src); err != nil {
				return err
			}
		}
		fmt.Printf("Wrote object: %s to bucket: %s\n", o+config.NameSuffix, config.DestinationBucket)
	}

	return nil
}