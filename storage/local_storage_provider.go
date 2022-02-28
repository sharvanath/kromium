package storage

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type LocalStorageProvider struct {}

func getFolderName(bucket string) string {
	return strings.TrimPrefix(bucket, "file://")
}

func (l LocalStorageProvider) ListObjects(ctx context.Context, bucket string) ([]string, error) {
	files, err := ioutil.ReadDir(getFolderName(bucket))
	if err != nil {
		return nil, err
	}

	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	return names, nil
}

func (l LocalStorageProvider) ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error) {
	f, err := os.Open(getFolderName(bucket) + "/" + object)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (g LocalStorageProvider) ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error) {
	f, err := os.OpenFile(getFolderName(bucket) + "/" + object, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (g LocalStorageProvider) DeleteObject(ctx context.Context, bucket string, object string) error {
	return os.Remove(getFolderName(bucket) + "/" + object)
}

func (g LocalStorageProvider) Close() error {
	return nil
}