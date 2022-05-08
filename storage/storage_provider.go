package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type S3Config struct {
	Region string
}

type StorageConfig struct {
	S3Config S3Config
}

type StorageProvider interface {
	// The caller will close.
	ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error)
	// The caller will close. Exception is that close should also flush any pending data.
	ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error)
	DeleteObject(ctx context.Context, bucket string, object string) error
	ListObjects(ctx context.Context, bucket string) ([]string, error)
	GetBucketName(ctx context.Context, bucketFullName string) (string, error)
	Close() error
}

func GetObjectWriter(ctx context.Context, s StorageProvider, bucket string, object string) (io.WriteCloser, error) {
	b, err := s.GetBucketName(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return s.ObjectWriter(ctx, b, object)
}

func GetObjectReader(ctx context.Context, s StorageProvider, bucket string, object string) (io.ReadCloser, error) {
	b, err := s.GetBucketName(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return s.ObjectReader(ctx, b, object)
}

func DeleteObject(ctx context.Context, s StorageProvider, bucket string, object string) error {
	b, err := s.GetBucketName(ctx, bucket)
	if err != nil {
		return err
	}
	return s.DeleteObject(ctx, b, object)
}

func ListObjects(ctx context.Context, s StorageProvider, bucket string) ([]string, error) {
	b, err := s.GetBucketName(ctx, bucket)
	if err != nil {
		return nil, err
	}
	return s.ListObjects(ctx, b)
}

func GetStorageProvider(ctx context.Context, uri string, storageConfig *StorageConfig) (StorageProvider, error) {
	if strings.HasPrefix(uri, "gs://") {
		return newGcsStorageProvider(ctx)
	}
	if strings.HasPrefix(uri, "file://") {
		return &LocalStorageProvider{}, nil
	}
	if strings.HasPrefix(uri, "s3://") {
		return newS3StorageProvider(storageConfig.S3Config.Region)
	}
	return nil, fmt.Errorf("No storage provider found for %s", uri)
}
