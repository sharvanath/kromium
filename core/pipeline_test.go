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
	os.Mkdir("/tmp/pipeline_test", 0700)
	src_dir = "/tmp/pipeline_test/" + randSeq(5)
	os.Mkdir(src_dir, 0700)
	dst_dir = src_dir + "_dst"
	os.Mkdir(dst_dir, 0700)
	state_dir = src_dir + "_state"
	os.Mkdir(state_dir, 0700)
	var files []string
	for i := 0; i < numFiles; i++ {
		file := fmt.Sprintf("%s/%d", src_dir, i)
		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
		if err != nil {
			fmt.Printf("Error in setup %s", err)
		}
		f.WriteString("test\n")
		files = append(files, file)
		f.Close()
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

func getFilesToMtime(dir string) (map[string]string, error) {
	dirObj, err := os.Open(dir)
	defer dirObj.Close()
	if err != nil {
		return nil, err
	}

	files, err := dirObj.Readdir(-1)
	if err != nil {
		return nil, err
	}

	fileToMtime := make(map[string]string)
	for _, f := range files {
		fileToMtime[f.Name()] = f.ModTime().String()
	}
	return fileToMtime, nil
}

func getKeyMap(filesToMTime map[string]string) map[string]bool {
	map1 := make(map[string]bool)
	for f, _ := range filesToMTime {
		map1[f] = true
	}
	return map1
}

func TestRunIdentityPipeline(t *testing.T) {
	setUp(10)
	defer tearDown()
	ctx := context.Background()
	RunPipelineLoop(ctx, getPipelineConfig(), 1, false)
	files, err := getFilesToMtime(src_dir)
	assert.NoError(t, err, "test error")
	filesDst, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	assert.Equal(t, getKeyMap(files), getKeyMap(filesDst))
}

func TestRunIdentitySuffixPipeline(t *testing.T) {
	setUp(10)
	defer tearDown()
	ctx := context.Background()
	config := getPipelineConfig()
	config.NameSuffix = "_test"
	RunPipelineLoop(ctx, config, 1, false)
	files, err := getFilesToMtime(src_dir)
	assert.NoError(t, err, "test error")
	filesDst, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	filesWSuffix := make(map[string]bool)
	for f, _ := range files {
		filesWSuffix[f+"_test"] = true
	}
	assert.Equal(t, filesWSuffix, getKeyMap(filesDst))
}

func TestSecondPipelineRunSkipsDone(t *testing.T) {
	setUp(10)
	defer tearDown()
	ctx := context.Background()
	config := getPipelineConfig()

	// first run
	RunPipeline(ctx, config, 0, false)
	filesDst, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")

	start_time := time.Now().String()
	// second run
	RunPipeline(ctx, config, 0, false)
	filesDst1, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	lastRunMtimes := make(map[string]string)
	assert.Equal(t, 10, len(filesDst1))

	for f, lastM := range filesDst1 {
		if _, ok := filesDst[f]; ok {
			lastRunMtimes[f] = lastM
		} else {
			assert.Greater(t, lastM, start_time)
		}
	}
	assert.Equal(t, filesDst, lastRunMtimes)
}

func TestCrashedPipelineRunIsReDone(t *testing.T) {
	setUp(10)
	defer tearDown()
	ctx := context.Background()
	config := getPipelineConfig()

	// first run
	RunPipeline(ctx, config, 0, false)
	filesDst, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	assert.Greater(t, len(filesDst), 1)

	// remove state files to simulate crash
	assert.NoError(t, os.RemoveAll(state_dir), "test error")
	os.Mkdir(state_dir, 0700)

	start_time := time.Now().String()
	// second run
	RunPipelineLoop(ctx, config, 1, false)
	filesDst1, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	assert.Equal(t, 10, len(filesDst1))

	for _, lastM := range filesDst1 {
		assert.Greater(t, lastM, start_time)
	}
}

func TestRunIdentityPipelineLargeSequential(t *testing.T) {
	setUp(1000)
	defer tearDown()
	ctx := context.Background()
	err := RunPipelineLoop(ctx, getPipelineConfig(), 1, false)
	assert.NoError(t, err, "error running pipeline")
	files, err := getFilesToMtime(src_dir)
	assert.NoError(t, err, "test error")
	filesDst, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	assert.Equal(t, getKeyMap(files), getKeyMap(filesDst))
}

func TestRunIdentityPipelineLargeParallel(t *testing.T) {
	setUp(5000)
	//defer tearDown()
	ctx := context.Background()
	err := RunPipelineLoop(ctx, getPipelineConfig(), 10, false)
	assert.NoError(t, err, "error running pipeline")
	files, err := getFilesToMtime(src_dir)
	assert.NoError(t, err, "test error")
	filesDst, err := getFilesToMtime(dst_dir)
	assert.NoError(t, err, "test error")
	assert.Equal(t, getKeyMap(files), getKeyMap(filesDst))
}