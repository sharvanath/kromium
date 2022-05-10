package storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"io"
	"strings"
)

type GcsStorageProvider struct {
	client *storage.Client
}

func newGcsStorageProvider(ctx context.Context) (StorageProvider, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GcsStorageProvider{ client }, nil
}

func getBucketName(bucket string) string {
	return strings.TrimPrefix(bucket, "gs://")
}

func (g GcsStorageProvider) Close() error {
	return g.client.Close()
}

func (g GcsStorageProvider) ListObjects(ctx context.Context, bucket string) ([]string, error) {
	query := &storage.Query{Prefix: ""}
	var names []string
	it := g.client.Bucket(getBucketName(bucket)).Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing bucket %s. %v", bucket, err)
		}
		names = append(names, attrs.Name)
	}
	return names, nil
}

// The caller must close
func (g GcsStorageProvider) ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error) {
	return g.client.Bucket(getBucketName(bucket)).Object(object).NewReader(ctx)
}

func (g GcsStorageProvider) ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error) {
	return g.client.Bucket(getBucketName(bucket)).Object(object).NewWriter(ctx), nil
}

func (g GcsStorageProvider) DeleteObject(ctx context.Context, bucket string, object string) error {
	return g.client.Bucket(getBucketName(bucket)).Object(object).Delete(ctx)
}

func (g GcsStorageProvider) GetBucketName(ctx context.Context, bucketFullName string) (string, error) {
	return strings.TrimPrefix(bucketFullName, "gs://"), nil
}