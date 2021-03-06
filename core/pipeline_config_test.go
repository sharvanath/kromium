package core

import (
	"context"
	"github.com/sharvanath/kromium/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getIdentityPipelineConfig(src string, dst string, state string) *PipelineConfig {
	config := PipelineConfig{}
	config.SourceBucket = "file://" + src
	config.DestinationBucket = "file://" + dst
	config.StateBucket = "file://" + state
	config.NameSuffix = ""

	var err error
	if config.sourceStorageProvider, err = storage.GetStorageProvider(context.Background(), config.SourceBucket, nil); err != nil {
		return nil
	}
	if config.destStorageProvider, err = storage.GetStorageProvider(context.Background(), config.DestinationBucket, nil); err != nil {
		return nil
	}
	if config.stateStorageProvider, err = storage.GetStorageProvider(context.Background(), config.StateBucket, nil); err != nil {
		return nil
	}
	identityTransform := TransformConfig{}
	identityTransform.Type = "Identity"
	config.Transforms = append(config.Transforms, identityTransform)
	config.Init(context.Background())
	return &config
}

func TestConfigHashNotEqualDifferentDestinationBucket(t *testing.T) {
	config1 := getIdentityPipelineConfig("a", "b", "c")
	config2 := *config1
	config2.DestinationBucket = "file:///tmp/dst_2"
	config2.Init(context.Background())
	assert.NotEqual(t, config1.getHash(), config2.getHash())
}

func TestConfigHashEqualDifferentStateBucket(t *testing.T) {
	config1 := getIdentityPipelineConfig("a", "b", "c")
	config2 := *config1
	config2.StateBucket = "file:///tmp/dst_2"
	config2.Init(context.Background())
	assert.Equal(t, config1.getHash(), config2.getHash())
}
