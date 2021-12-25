package core

import (
	"context"
	"fmt"
	"encoding/gob"
	"github.com/sharvanath/kromium/storage"
	"math/rand"
	"strings"
	"time"
)

type WorkerState struct {
	// Only the bitmap is serialized
	bitmap []byte
	// in-memory state
	// batch size is 8
	workerId string
	transformHash string
	mergedFiles []string
}

func createState(config *PipelineConfig, numFiles int)  *WorkerState {
	var w WorkerState
	bitmap_size := numFiles >> 3 + 1
	w.bitmap = make([]byte, bitmap_size)
	w.transformHash = config.getHash()
	return &w
}

func (w *WorkerState) mergeState(w2 WorkerState) error {
	if len(w.bitmap) != len(w2.bitmap) {
		return fmt.Errorf("inconsistent bitmap lengths during state merge: %d %d", len(w.bitmap), len(w2.bitmap))
	}

	for i, byte := range w.bitmap {
		w.bitmap[i] = byte | w2.bitmap[i]
	}
	return nil
}

func (w *WorkerState) findRandomEmpty()  int {
	bitmapSize := len(w.bitmap)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Int() % bitmapSize
	for i := 0 ; i < bitmapSize; i++ {
		idx := (n + i) % bitmapSize
		b := w.bitmap[idx]
		if b == 0 {
			return idx << 3
		}
	}

	return -1
}

func (w WorkerState) fileName() string {
	return w.transformHash + "_" + sha1Str(w.workerId)
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
			fmt.Printf("Ignoring state file %s not matching transform hash %s", f, pipeline.getHash())
			continue
		}
		reader, err := stateStorageProvider.ObjectReader(ctx, pipeline.StateBucket, f)
		if err != nil {
			return nil, err
		}
		var currState WorkerState
		if err = gob.NewDecoder(reader).Decode(&currState.bitmap); err != nil {
			fmt.Printf("Could not decode worker file %s %s", f, err)
			return nil, err
		}
		err = w.mergeState(currState)
		if err != nil {
			return nil, err
		}
		w.mergedFiles = append(w.mergedFiles, f)
	}

	return w, nil
}

func WriteState(ctx context.Context, stateBucket string, w *WorkerState) (error) {
	stateStorageProvider := storage.GetStorageProvider(stateBucket)
	writer, err := stateStorageProvider.ObjectWriter(ctx, stateBucket, w.fileName())
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(writer)
	enc.Encode(w.bitmap)

	for _, f := range w.mergedFiles {
		// Ignore errors during delete since the object might be already deleted
		stateStorageProvider.DeleteObject(ctx, stateBucket, f)
	}
	return nil
}