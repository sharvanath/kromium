package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getIdentityPipelineConfig(src string, dst string, state string) *PipelineConfig {
	config := PipelineConfig{}
	config.SourceBucket = "file://" + src
	config.DestinationBucket = "file://" + dst
	config.StateBucket = "file://" + state
	config.NameSuffix = ""

	identityTransform := TransformConfig{}
	identityTransform.Name = "Identity"
	config.Transforms = append(config.Transforms, identityTransform)

	return &config
}

func TestConfigHashNotEqualDifferentDestinationBucket(t *testing.T) {
	config1 := getIdentityPipelineConfig("a", "b", "c")
	config2 := *config1
	config2.DestinationBucket = "file:///tmp/dst_2"
	assert.NotEqual(t, config1.getHash(), config2.getHash())
}

func TestConfigHashEqualDifferentStateBucket(t *testing.T) {
	config1 := getIdentityPipelineConfig("a", "b", "c")
	config2 := *config1
	config2.StateBucket = "file:///tmp/dst_2"
	assert.Equal(t, config1.getHash(), config2.getHash())
}
