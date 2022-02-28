package core

import (
	"context"
	"encoding/json"
	"github.com/sharvanath/kromium/storage"
	"os"
)

type TransformConfig struct {
	Name string
	Args interface{}
}

type PipelineConfig struct {
	SourceBucket      string
	DestinationBucket string
	StateBucket       string
	NameSuffix        string
	Transforms        []TransformConfig
	Hash              string

	// Transient fields
	sourceStorageProvider storage.StorageProvider
	destStorageProvider storage.StorageProvider
	stateStorageProvider storage.StorageProvider
}

func (p PipelineConfig) getHash() string {
	if len(p.Hash) != 0 {
		return p.Hash
	}
	h := newSha1Hasher()
	h.addStr(p.SourceBucket)
	h.addStr(p.DestinationBucket)
	h.addStr(p.NameSuffix)
	for _, t := range p.Transforms {
		h.addStr(t.Name)
	}
	p.Hash = h.getStrHash()
	return p.Hash
}

func ReadPipelineConfigFile(ctx context.Context, configFile string) (*PipelineConfig, error) {
	file, _ := os.Open(configFile)
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := PipelineConfig{}
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	inputStorageProvider, err := storage.GetStorageProvider(ctx, config.SourceBucket)
	if err != nil {
		return nil, err
	}
	config.sourceStorageProvider = inputStorageProvider

	outputStorageProvider, err := storage.GetStorageProvider(ctx, config.DestinationBucket)
	if err != nil {
		return nil, err
	}
	config.destStorageProvider = outputStorageProvider

	stateStorageProvider, err := storage.GetStorageProvider(ctx, config.StateBucket)
	if err != nil {
		return nil, err
	}
	config.stateStorageProvider = stateStorageProvider

	return &config, nil
}

func (p PipelineConfig) Close() error {
	if p.sourceStorageProvider != nil {
		if err := p.sourceStorageProvider.Close(); err != nil {
			return err
		}
	}

	if p.destStorageProvider != nil {
		if err := p.destStorageProvider.Close(); err != nil {
			return err
		}
	}

	if p.stateStorageProvider != nil {
		if err := p.stateStorageProvider.Close(); err != nil {
			return err
		}
	}

	return nil
}
