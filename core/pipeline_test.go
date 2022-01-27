package core

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

func getNewPipelineConfigFromStr() string {
	//ecoder := json.NewDecoder(file)
	//config := core.PipelineConfig{}
	//err := decoder.Decode(&config)
	return ""
}

func TestSingleProcessor(t *testing.T) {
	assert.Equal(t, "2f3a624b8f251c930571a35b1932d9e6e9dc2da1", sha1Str("sharva"))
}
