package core

import (
	"encoding/json"
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

func ReadPipelineConfigFile(configFile string) (*PipelineConfig, error) {
	file, _ := os.Open(configFile)
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := PipelineConfig{}
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
