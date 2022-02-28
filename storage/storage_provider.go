package storage

import (
	"context"
	"io"
	"strings"
)

type StorageProvider interface {
	// The caller will close.
	ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error)
	// The caller will close.
	ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error)
	DeleteObject(ctx context.Context, bucket string, object string) error
	ListObjects(ctx context.Context, bucket string) ([]string, error)
	Close() error
}

func GetStorageProvider(ctx context.Context, uri string) (StorageProvider, error) {
	if strings.HasPrefix(uri, "gs://") {
		return newGcsStorageProvider(ctx)
	}
	if strings.HasPrefix(uri, "file://") {
		return &LocalStorageProvider{}, nil
	}
	return nil, nil
}
