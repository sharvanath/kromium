package storage

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"io/ioutil"
)

const DEFAULT_FILE_BUF_SIZE_BYTES int = 10*1024

type S3StorageProvider struct {
	session *session.Session
}

func newS3StorageProvide() (StorageProvider, error) {
	// The session the S3 Downloader will use
	session := session.Must(session.NewSession())
	return &S3StorageProvider{session: session}, nil
}

// The caller must close
func (s S3StorageProvider) ObjectReader(ctx context.Context, bucket string, object string) (io.ReadCloser, error) {
	// Create a downloader with the session and default options
	downloader := s3manager.NewDownloader(s.session)
	w := aws.NewWriteAtBuffer(make([]byte, DEFAULT_FILE_BUF_SIZE_BYTES))
	_, err := downloader.Download(w, &s3.GetObjectInput{
		Bucket: &bucket,
		Key: &object,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file, %v", err)
	}
	return ioutil.NopCloser(bytes.NewReader(w.Bytes())), nil
}

func (s S3StorageProvider) Close() error {
	return nil
}

type S3ObjectWriter struct {
	b bytes.Buffer
	w io.Writer
	s *S3StorageProvider
	bucket, object string
}

func (o S3ObjectWriter) Close() error {
	uploader := s3manager.NewUploader(o.s.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: &o.bucket,
		Key: &o.object,
		Body: bufio.NewReader(&o.b),
	})
	if err != nil {
		return err
	}
	return nil
}

func (o S3ObjectWriter) Write(p []byte) (n int, err error) {
	return o.w.Write(p)
}

func (s S3StorageProvider) ObjectWriter(ctx context.Context, bucket string, object string) (io.WriteCloser, error) {
	var o S3ObjectWriter
	o.w = bufio.NewWriter(&o.b)
	o.bucket = bucket
	o.object = object
	return &o, nil
}

func (s S3StorageProvider) DeleteObject(ctx context.Context, bucket string, object string) error {
	svc := s3.New(s.session)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key: &object,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s S3StorageProvider) ListObjects(ctx context.Context, bucket string) ([]string, error) {
	svc := s3.New(s.session)
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &bucket})
	if err != nil {
		return nil, err
	}

	var objects []string
	for _, f := range resp.Contents {
		objects = append(objects, *f.Key)
	}
	return objects, nil
}