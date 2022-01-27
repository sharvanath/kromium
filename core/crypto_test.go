package core

import "testing"
import "github.com/stretchr/testify/assert"

func TestSha1Sum(t *testing.T) {
	assert.Equal(t, "2f3a624b8f251c930571a35b1932d9e6e9dc2da1", sha1Str("sharva"))
}

func TestSha1Hasher(t *testing.T) {
	h := newSha1Hasher()
	h.addStr("sharva")
	assert.Equal(t, "2f3a624b8f251c930571a35b1932d9e6e9dc2da1", h.getStrHash())
}

func TestSha1HasherAddStrMult(t *testing.T) {
	h := newSha1Hasher()
	h.addStr("sha")
	h.addStr("rva")
	assert.Equal(t, "2f3a624b8f251c930571a35b1932d9e6e9dc2da1", h.getStrHash())
}

func TestSha1HasherAddStrSlice(t *testing.T) {
	h := newSha1Hasher()
	x := []string{"sha", "rva"}
	h.addStrSlice(x)
	assert.Equal(t, "2f3a624b8f251c930571a35b1932d9e6e9dc2da1", h.getStrHash())
}