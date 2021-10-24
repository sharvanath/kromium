package core

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"io"
	"log"
)

func listFiles(ctx context.Context, bucket string) ([]string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	query := &storage.Query{Prefix: ""}
	var names []string
	it := client.Bucket(bucket).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		names = append(names, attrs.Name)
	}
	return names, nil
}

func objectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucket).Object(object).NewReader(ctx)
}

func objectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(bucket).Object(object).NewWriter(ctx), nil
}

func RunPipeline(ctx context.Context, config PipelineConfig) error {
	files, err := listFiles(ctx, config.SourceBucket)
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
				srcCloser, err := objectReader(ctx, config.SourceBucket, o)
				defer srcCloser.Close()
				src = srcCloser
				if err != nil {
					return  err
				}
			} else {
				src = bufio.NewReader(buf)
			}

			if idx == len(config.Transforms) - 1 {
				dstCloser, err := objectWriter(ctx, config.DestinationBucket, o+config.NameSuffix)
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

			if _, err := getTransform(t.Name).Transform(dst, src); err != nil {
				return err
			}
		}
		fmt.Printf("Wrote object: %s to bucket: %s\n", o+config.NameSuffix, config.DestinationBucket)
	}

	return nil
}