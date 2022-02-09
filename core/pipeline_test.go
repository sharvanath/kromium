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

func getFiles(dir string) ([]string, error) {
	dirObj, err := os.Open(dir)
	if err != nil {
		return nil, err
	}

	files, err := dirObj.Readdir(1000)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	return fileNames, nil
}

func sliceToMap(slice []string) map[string]bool {
	map1 := make(map[string]bool)
	for _, f := range slice {
		map1[f] = true
	}
	return map1
}

func TestRunIdentityPipeline(t *testing.T) {
	ctx := context.Background()
	RunPipelineLoop(ctx, getPipelineConfig())
	files, err := getFiles(src_dir)
	assert.NoError(t, err, "error running pipeline")
	filesDst, err := getFiles(dst_dir)
	assert.NoError(t, err, "error running pipeline")
	assert.Equal(t, files, filesDst)
}

func TestRunIdentitySuffixPipeline(t *testing.T) {
	ctx := context.Background()
	config := getPipelineConfig()
	config.NameSuffix = "_test"
	RunPipelineLoop(ctx, config)
	files, err := getFiles(src_dir)
	assert.NoError(t, err, "error running pipeline")
	filesDst, err := getFiles(dst_dir)
	assert.NoError(t, err, "error running pipeline")
	var filesWSuffix []string
	for _, f := range files {
		filesWSuffix = append(filesWSuffix, f + "_test")
	}
	assert.Equal(t, sliceToMap(filesWSuffix), sliceToMap(filesDst))
}

func TestSecondPipelineRunSkipsDone(t *testing.T) {
}

func TestCrashedPipelineRunIsReDone(t *testing.T) {
}

func TestParallelWorkers(t *testing.T) {
}
