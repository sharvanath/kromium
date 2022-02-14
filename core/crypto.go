package core

import (
	"crypto/sha1"
	"encoding/hex"
	"hash"
)

func sha1Str(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

type Hasher struct {
	h hash.Hash
}

func newSha1Hasher() Hasher {
	return Hasher{sha1.New()}
}

func (h *Hasher) addStr(input string) {
	h.h.Write([]byte(input))
}

func (h *Hasher) addStrSlice(input []string) {
	for _, i := range input {
		h.h.Write([]byte(i))
	}
}

func (h *Hasher) getStrHash() string {
	return hex.EncodeToString(h.h.Sum(nil))
}
