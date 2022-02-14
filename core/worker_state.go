package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/sharvanath/kromium/storage"
	"os"
	"strings"
	log "github.com/sirupsen/logrus"
)

// The batch size
const cBatchSize = 10

// Only the byte slice is serialized to the state file. workerId is used for the state file name.
type WorkerState struct {
	// One bit for each batch. If the bit is 1 that means the batch has been processed already.
	m *bitmap
	// The following is just in-memory state
	numFiles      int
	workerId      string
	transformHash string
	mergedFiles   []string
}

func createState(config *PipelineConfig, numFiles int) *WorkerState {
	var w WorkerState
	// Add an extra partial batch in case numFile is not perfectly divisible.
	b_size := numFiles / cBatchSize
	if numFiles % cBatchSize != 0 {
		b_size += 1
	}

	w.numFiles = numFiles
	w.m = newBitmap(b_size)
	w.transformHash = config.getHash()
	return &w
}

func (w *WorkerState) mergeState(w2 WorkerState) error {
	if w.m.size != w2.m.size {
		return fmt.Errorf("inconsistent bitmap lengths during state merge: %d %d", w.m.size, w2.m.size)
	}

	w.m.merge(w2.m)
	return nil
}

func (w *WorkerState) fileName() string {
	return w.transformHash + "_" + sha1Str(w.workerId)
}

// Process files [start, end)
func (w *WorkerState) findProcessingRange() (int, int) {
	idx := w.m.findRandomEmpty()
	if idx == -1 {
		return -1, -1
	}
	end := idx * cBatchSize + cBatchSize
	if end > w.numFiles {
		end = w.numFiles
	}
	return idx * cBatchSize, end
}

func (w *WorkerState) setProcessed(startOff int) {
	batchIdx := startOff / cBatchSize
	w.m.set(batchIdx)
}

func ReadMergedState(ctx context.Context, pipeline *PipelineConfig, numFiles int) (*WorkerState, error) {
	stateStorageProvider := storage.GetStorageProvider(pipeline.StateBucket)
	files, err := stateStorageProvider.ListObjects(ctx, pipeline.StateBucket)
	if err != nil {
		return nil, err
	}

	w := createState(pipeline, numFiles)

	for _, f := range files {
		if !strings.HasPrefix(f, pipeline.getHash()) {
			log.Warnf("Ignoring state file %s not matching transform hash %s", f, pipeline.getHash())
			continue
		}
		reader, err := stateStorageProvider.ObjectReader(ctx, pipeline.StateBucket, f)
		// The file could be deleted by the time we get to it.
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		if err == nil {
			var currState WorkerState
			m, err := readFrom(reader)
			reader.Close()
			if err != nil {
				// ignore errors since these could happen due to concurrent deletes
				// worst case this leads to duplicate work
				log.Infof("Could not decode worker file %s %s", f, err)
				continue
			}
			currState.m = m
			err = w.mergeState(currState)
			if err != nil {
				// ignore errors since these could happen due to concurrent deletes
				// worst case this leads to duplicate work
				log.Infof("Could not decode worker file %s %s", f, err)
				continue
			}
		}
		w.mergedFiles = append(w.mergedFiles, f)
	}

	return w, nil
}

func WriteState(ctx context.Context, stateBucket string, w *WorkerState) error {
	stateStorageProvider := storage.GetStorageProvider(stateBucket)
	writer, err := stateStorageProvider.ObjectWriter(ctx, stateBucket, w.fileName())
	if err != nil {
		return err
	}
	defer writer.Close()
	w.m.writeTo(writer)
	for _, f := range w.mergedFiles {
		// Ignore errors during delete since the object might be already deleted
		stateStorageProvider.DeleteObject(ctx, stateBucket, f)
	}
	return nil
}
