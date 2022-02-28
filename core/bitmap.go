package core

import (
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"time"
)

var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))

type bitmap struct {
	slice []byte
	// Note that the last byte could be partially used if
	// size is not a multiple of 8
	size int
}

func newBitmap(size int) *bitmap {
	var m bitmap
	m.size = size
	sliceLen := size >> 3
	if size%8 != 0 {
		sliceLen += 1
	}
	m.slice = make([]byte, sliceLen)
	// Set the unused bits to 1 so that the
	// search for empty bit is easier.
	if size%8 != 0 {
		m.slice[len(m.slice)-1] = ^(1<<(size%8) - 1)
	}
	return &m
}

func hasEmptyBit(x byte) bool {
	return x != byte(0xff)
}

// 0 for LSB, and 7 for MSB
func returnEmptyBit(x byte) int {
	var idx []int
	for i := 0; i < 8; i += 1 {
		if x&1 == 0 {
			idx = append(idx, i)
		}
		x = x >> 1
	}
	if len(idx) == 0 {
		return -1
	}

	return idx[randomGenerator.Int() % len(idx)]
}

func countSetBits(n byte) int {
	count := 0
	for ;n > 0; {
		n &= n - 1
		count++
	}
	return count
}

func (m *bitmap) set(idx int) {
	if idx < 0 || idx >= m.size {
		log.Fatalf("Bad index %d", idx)
	}

	// byte0: 0LSB, 7MSB, byte1: 8LSB, 15MSB
	m.slice[idx>>3] |= 1 << (idx % 8)
}

func (m *bitmap) clear(idx int) {
	if idx < 0 || idx >= m.size {
		log.Fatalf("Bad index %d", idx)
	}

	// byte0: 0LSB, 7MSB, byte1: 8LSB, 15MSB
	m.slice[idx>>3] &= ^(1 << (idx % 8))
}

// Finds a random bit that is free. Note that for the last byte we mark all the extra unused bits as 1 already
// so we need not worry about that part.
func (m *bitmap) findRandomEmpty() int {
	n := randomGenerator.Int() % len(m.slice)
	for i := 0; i < len(m.slice); i++ {
		idx := (n + i) % len(m.slice)
		b := m.slice[idx]
		if hasEmptyBit(b) {
			offset := returnEmptyBit(b)
			if offset == -1 {
				log.Fatal("Illegal state, empty bit not found")
			}
			return idx<<3 + returnEmptyBit(b)
		}
	}
	return -1
}

func (m *bitmap) merge(m1 *bitmap) {
	for i := 0; i < len(m.slice); i += 1 {
		m.slice[i] |= m1.slice[i]
	}
}

func (m *bitmap) writeTo(writer io.Writer) error {
	encoder := gob.NewEncoder(writer)
	if err := encoder.Encode(m.slice); err != nil {
		return err
	}
	if err := encoder.Encode(m.size); err != nil {
		return err
	}
	return nil
}

func (m *bitmap) usedSize() int {
	c := 0
	for i := 0; i < len(m.slice); i++ {
		b := m.slice[i]
		if b != 0 {
			c += countSetBits(b)
		}
	}
	return c
}

func readFrom(reader io.Reader) (*bitmap, error) {
	decoder := gob.NewDecoder(reader)
	var m bitmap
	if err := decoder.Decode(&m.slice); err != nil {
		return nil, err
	}
	if err := decoder.Decode(&m.size); err != nil {
		return nil, err
	}
	return &m, nil
}