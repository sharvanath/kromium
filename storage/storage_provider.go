package storage

import (
	"context"
	"io"
	"strings"
)

type StorageProvider interface {
	ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error)
	ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error)
	ListObjects(ctx context.Context, bucket string) ([]string, error)
}

func GetStorageProvider(uri string) StorageProvider {
	if strings.HasPrefix(uri, "gs://") {
		return GcsStorageProvider{}
	}
	return nil
}
