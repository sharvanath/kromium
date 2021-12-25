package transforms

import (
	"io"
)

type Transform interface {
	// The Transformation to apply on the reader and output to the writer.
	// Optionally returns any conversion metadata and error
	Transform(dst io.Writer, src io.Reader) (interface{}, error)
}

func GetTransform(name string, args interface{}) Transform {
	switch name {
	case "Identity":
		return IdentityTransform{}
	case "GzipCompress":
		return GzipCompressTransform{args.(map[string]interface{})}
	case "GzipDecompress":
		return GzipDecompressTransform{}
	case "SedTransform":
		return SedTransform{args.(string)}
	case "DecryptionTransform":
		return NewDecryptionTransform(args.(map[string]interface{}))
	case "EncryptionTransform":
		return NewEncryptionTransform(args.(map[string]interface{}))
	}
	return nil
}