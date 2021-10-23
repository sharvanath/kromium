package core

import (
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

func readObject(ctx context.Context, bucket string, object string) ([]byte, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	defer reader.Close()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024*16)
	var input []byte
	bytes := 0
	n := 0
	for n, err := reader.Read(buf); err == nil; n, err = reader.Read(buf) {
		input = append(input, buf...)
		fmt.Printf("Read %d total %d from object: %s, bucket: %s\n", n, bytes, object, bucket)
		bytes += n
	}
	if n != 0 {
		input = append(input, buf[:n]...)
	}
	if err == io.EOF {
		fmt.Println("End")
		return input, nil
	}

	return input, err
}

func writeObject(ctx context.Context, bucket string, object string, buf []byte) (error) {
	fmt.Printf("Writing to %d to bucket: %s, object: %s\n", len(buf), bucket, object)
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	writer := client.Bucket(bucket).Object(object).NewWriter(ctx)
	n, err := writer.Write(buf)
	if n != len(buf) {
		return fmt.Errorf("Only wrote %d while neeed %d, for bucket: %s, object:  %s", n, len(buf), bucket, object)
	}
	writer.Close()
	return err
}

func RunSync(ctx context.Context, config SyncConfig) error {
	files, err := listFiles(ctx, config.SourceBucket)
	if err != nil {
		return err
	}

	for _, o := range files {
		for _, t := range config.Operations {
			fmt.Printf("Processing object: %s from bucket: %s\n", config.SourceBucket, o)
			inputBuf, err := readObject(ctx, config.SourceBucket, o)
			if err != nil {
				return  err
			}
			fmt.Println(len(inputBuf))
			ouputBuf := getTransform(t).transform(inputBuf)
			writeObject(ctx, config.DestinationBucket, o+config.NameSuffix, ouputBuf)
		}
	}

	return nil
}