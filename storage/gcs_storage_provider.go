package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"google.golang.org/api/iterator"
	"io"
	"log"
	"strings"
)

type GcsStorageProvider struct {}

func getBucketName(bucket string) string {
	return strings.TrimPrefix(bucket, "gs://")
}

func (g GcsStorageProvider) ListObjects(ctx context.Context, bucket string) ([]string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	query := &storage.Query{Prefix: ""}
	var names []string
	it := client.Bucket(getBucketName(bucket)).Objects(ctx, query)
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

func (g GcsStorageProvider) ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(getBucketName(bucket)).Object(object).NewReader(ctx)
}

func (g GcsStorageProvider) ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(getBucketName(bucket)).Object(object).NewWriter(ctx), nil
}

func (g GcsStorageProvider) DeleteObject(ctx context.Context, bucket string, object string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	return client.Bucket(getBucketName(bucket)).Object(object).Delete(ctx)
}