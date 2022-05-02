package core

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

// The batch size
const cBatchSize = 16

// Only the byte slice is serialized to the state file. workerId is used for the state file name.
type WorkerState struct {
	// One bit for each batch. If the bit is 1 that means the batch has been processed already.
	m *bitmap
	// The following is just in-memory state
	numFiles      int
	processed     int
	workerId      string
	mergedFiles   []string
	pipeline      *PipelineConfig
}

func createState(pipeline *PipelineConfig, numFiles int) *WorkerState {
	var w WorkerState
	// Add an extra partial batch in case numFile is not perfectly divisible.
	b_size := numFiles / cBatchSize
	if numFiles%cBatchSize != 0 {
		b_size += 1
	}

	w.pipeline = pipeline
	w.numFiles = numFiles
	w.m = newBitmap(b_size)
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
	return w.pipeline.getHash() + "_" + sha1Str(w.workerId)
}

// Process files [start, end)
func (w *WorkerState) findProcessingRange() (int, int) {
	idx := w.m.findRandomEmpty()
	if idx == -1 {
		return -1, -1
	}
	end := idx*cBatchSize + cBatchSize
	if end > w.numFiles {
		end = w.numFiles
	}
	return idx * cBatchSize, end
}

func (w *WorkerState) setProcessed(startOff int) {
	batchIdx := startOff / cBatchSize
	w.m.set(batchIdx)
}

type WorkerStateResp struct {
	w *WorkerState
	e error
}
func ReadMergedState(ctx context.Context, pipeline *PipelineConfig, numFiles int) (*WorkerState, error) {
	files, err := pipeline.stateStorageProvider.ListObjects(ctx, pipeline.StateBucket)
	if err != nil {
		return nil, err
	}

	w := createState(pipeline, numFiles)

	var channels []chan WorkerStateResp
	for _, f := range files {
		channel := make(chan WorkerStateResp)
		channels = append(channels, channel)
		go func(file string) {
			var w WorkerStateResp
			if !strings.HasPrefix(file, pipeline.getHash()) {
				log.Warnf("Ignoring state file %s not matching transform hash %s", file, pipeline.getHash())
				w.e = fmt.Errorf("ignoring state file %s", file)
				channel <- w
				return
			}
			reader, err := pipeline.stateStorageProvider.ObjectReader(ctx, pipeline.StateBucket, file)
			// The file could be deleted by the time we get to it.
			if err != nil {
				w.e = err
				channel <- w
				return
			}

			if err == nil {
				var currState WorkerState
				currState.pipeline = pipeline
				m, err := readFrom(reader)
				reader.Close()
				if err != nil {
					// ignore errors since these could happen due to concurrent deletes
					// worst case this leads to duplicate work
					log.Infof("Could not decode worker file %s %s", f, err)
					w.e = err
					channel <- w
					return
				}
				w.w = &currState
				currState.m = m
			}
			channel <- w
			return
		}(f)
	}

	for i, c := range channels {
		stateResp := <- c
		if stateResp.e != nil {
			continue
		}
		if stateResp.w == nil {
			panic("Unexpected null worker state")
		}
		err = w.mergeState(*stateResp.w)
		if err != nil {
			// Ignore corrupt state files
			log.Errorf("corrupt state file %s %s", files[i], err)
			continue
		}
		w.mergedFiles = append(w.mergedFiles, files[i])
	}

	return w, nil
}

func WriteState(ctx context.Context, stateBucket string, w *WorkerState) error {
	c := make(chan error)
	go func() {
		writer, err := w.pipeline.stateStorageProvider.ObjectWriter(ctx, stateBucket, w.fileName())
		if err != nil {
			log.Debugf("Error in Writing %s %v", w.fileName(), err)
			c <- err
			return
		}
		w.m.writeTo(writer)
		writer.Close()
		c <- nil
	}()

	var wg sync.WaitGroup
	for _, f := range w.mergedFiles {
		wg.Add(1)
		go func(file string) {
			// Ignore errors during delete since the object might be already deleted
			err := w.pipeline.stateStorageProvider.DeleteObject(ctx, stateBucket, file)
			if err != nil {
				log.Debugf("Error in deleting %s %v", f, err)
			}
			wg.Done()
		}(f)
	}
	wg.Wait()
	err:= <- c
	return err
}
