package core

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var src_dir, dst_dir, state_dir string

func setUp(numFiles int) {
	rand.Seed(time.Now().UnixNano())
	src_dir = "/tmp/" + randSeq(5)
	os.Mkdir(src_dir, 0700)
	dst_dir = src_dir + "_dst"
	os.Mkdir(dst_dir, 0700)
	state_dir = src_dir + "_state"
	os.Mkdir(state_dir, 0700)
	var files []string
	for i := 0; i < numFiles; i++ {
		file := fmt.Sprintf("%s/%d", src_dir, i)
		os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0400)
		files = append(files, file)
	}
}

func tearDown() {
	os.RemoveAll(src_dir)
	os.RemoveAll(dst_dir)
	os.RemoveAll(state_dir)
}

func getPipelineConfig() *PipelineConfig {
	return getIdentityPipelineConfig(src_dir, dst_dir, state_dir)
}

func TestMain(m *testing.M) {
	setUp(10)
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestRunIdentityPipeline(t *testing.T) {
	ctx := context.Background()
	err := RunPipeline(ctx, getPipelineConfig())
	assert.NoError(t, err, "error running pipeline")
}