package schema

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestValidExampleConfigs(t *testing.T) {
	f, err := os.Open("../examples")
	assert.NoError(t, err, "test error reading examples dir")
	in, err := f.Readdir(-1)
	assert.NoError(t, err, "test error reading examples dir")
	for _, f := range in {
		err = ValidatePipelineConfig("../examples/" + f.Name())
		assert.NoError(t, err, "test error reading examples dir")
	}
}
