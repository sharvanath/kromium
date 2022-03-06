package core

import (
	"context"
	"github.com/sharvanath/kromium/storage"
	"log"
)

type TransformConfig struct {
	Type string
	Args interface{}
}

type PipelineConfig struct {
	SourceBucket      string
	DestinationBucket string
	StateBucket       string
	NameSuffix        string
	Transforms        []TransformConfig

	// Derived fields
	Hash              string
	sourceStorageProvider storage.StorageProvider
	destStorageProvider storage.StorageProvider
	stateStorageProvider storage.StorageProvider
}

func (p *PipelineConfig) getHash() string {
	if len(p.Hash) == 0 {
		log.Fatal("Hash should be populated at the beginning")
	}
	return p.Hash
}

func (p *PipelineConfig) Init(ctx context.Context) error {
	inputStorageProvider, err := storage.GetStorageProvider(ctx, p.SourceBucket)
	if err != nil {
		return err
	}
	p.sourceStorageProvider = inputStorageProvider

	outputStorageProvider, err := storage.GetStorageProvider(ctx, p.DestinationBucket)
	if err != nil {
		return err
	}
	p.destStorageProvider = outputStorageProvider

	stateStorageProvider, err := storage.GetStorageProvider(ctx, p.StateBucket)
	if err != nil {
		return err
	}
	p.stateStorageProvider = stateStorageProvider

	h := newSha1Hasher()
	h.addStr(p.SourceBucket)
	h.addStr(p.DestinationBucket)
	h.addStr(p.NameSuffix)
	for _, t := range p.Transforms {
		h.addStr(t.Type)
	}
	p.Hash = h.getStrHash()
	return nil
}

func (p *PipelineConfig) Close() error {
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
