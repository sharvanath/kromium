package core

import (
	"github.com/sharvanath/kromium/transforms"
	"io"
)

type Transform interface {
	// The Transformation to apply on the reader and output to the writer.
	// Optionally returns any conversion metadata and error
	Transform(dst io.Writer, src io.Reader) (interface{}, error)
}

func getTransform(name string) Transform {
	if name == "Identity" {
		return transforms.IdentityTransform{}
	}
	return nil
}