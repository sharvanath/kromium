package core

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCreateBitMap(t *testing.T) {
	b := newBitmap(8)
	assert.Equal(t, 1, len(b.slice))

	b = newBitmap(15)
	assert.Equal(t, 2, len(b.slice))
	// Size only 7 bits are used, the last one
	// should be set at initialization
	assert.Equal(t, byte(1<<7), b.slice[1])
}

func TestSet(t *testing.T) {
	b := newBitmap(20)
	b.set(9)
	assert.Equal(t, byte(0), b.slice[0])
	assert.Equal(t, byte(2), b.slice[1])
}

func TestFindEmptyBit(t *testing.T) {
	b := newBitmap(2)
	b.set(0)
	assert.Equal(t, 1, b.findRandomEmpty())
}

func TestFindSingleEmptyBit(t *testing.T) {
	b := newBitmap(21)
	for i := 0; i < 21; i += 1 {
		b.set(i)
	}
	b.clear(9)
	assert.Equal(t, 9, b.findRandomEmpty())
}

func TestFindNoEmptyBit(t *testing.T) {
	b := newBitmap(21)
	for i := 0; i < 21; i += 1 {
		b.set(i)
	}
	assert.Equal(t, -1, b.findRandomEmpty())
}

func TestMerge(t *testing.T) {
	b := newBitmap(21)
	b1 := newBitmap(21)
	b1.set(8)
	b.set(9)
	assert.Equal(t, byte(2), b.slice[1])
	assert.Equal(t, byte(1), b1.slice[1])
	b.merge(b1)
	assert.Equal(t, byte(3), b.slice[1])
}

func TestBitMapSerialization(t *testing.T) {
	b := newBitmap(8)
	b.set(1)
	assert.Equal(t, byte(2), b.slice[0])
	f, err := os.OpenFile("/tmp/bitmap_test_file", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	assert.NoError(t, err, "test error")
	assert.NoError(t, b.writeTo(f), "error serializing")
	f.Close()

	f1, err := os.OpenFile("/tmp/bitmap_test_file", os.O_RDONLY, 0700)
	assert.NoError(t, err, "test error")
	b1, err := readFrom(f1)
	assert.NoError(t, err, "error deserializaing")
	assert.Equal(t, byte(2), b1.slice[0])
	os.Remove("/tmp/bitmap_test_file")
}
