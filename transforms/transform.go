package transforms

import (
	"io"
)

type Transform interface {
	// The Transformation to apply on the reader and output to the writer.
	// Optionally returns any conversion metadata and error.
	// The Transform runs on the full src and writes to dst.
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
	case "Sed":
		return SedTransform{args.(string)}
	case "Decrypt":
		return NewDecryptionTransform(args.(map[string]interface{}))
	case "Encrypt":
		return NewEncryptionTransform(args.(map[string]interface{}))
	}
	return nil
}