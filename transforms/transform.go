package transforms

import (
	"io"
)

type Transform interface {
	// The Transformation to apply on the reader and output to the writer.
	// Optionally returns any conversion metadata and error
	Transform(dst io.Writer, src io.Reader) (interface{}, error)
}

func GetTransform(name string, _ interface{}) Transform {
	switch name {
	case "Identity":
		return IdentityTransform{}
	case "GzipCompress":
		return GzipCompressTransform{}
	case "GzipDecompress":
		return GzipDecompressTransform{}
	}
	return nil
}