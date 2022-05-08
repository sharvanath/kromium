// +build integration

// Run the tests with -tags=integration
package storage

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestS3Read(t *testing.T) {
	s, err := newS3StorageProvider("us-east-1")
	assert.NoError(t, err)
	r, err := s.ObjectReader(context.Background(), "kromium-src", "hello")
	assert.NoError(t, err)
	b := make([]byte, 40)
	size, err := r.Read(b)
	fmt.Printf("Read: %s, size: %d\n", string(b), size)
	r.Close()
}

func TestS3Write(t *testing.T) {
	s, err := newS3StorageProvider("us-east-1")
	assert.NoError(t, err)
	w, err := s.ObjectWriter(context.Background(), "kromium-src", "tmp1")
	assert.NoError(t, err)
	written, err := w.Write([]byte("Hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, written)
	assert.NoError(t, w.Close())
}
