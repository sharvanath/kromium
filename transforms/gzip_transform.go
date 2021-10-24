package transforms

import (
	"compress/gzip"
	"io"
)

type GzipCompressConfig struct {
	level int
}

type GzipCompressTransform struct {
	args map[string]interface{}
}

type GzipDecompressTransform struct {
}

func (i GzipCompressTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	compressWriter := gzip.NewWriter(dst)
	_, err := io.Copy(compressWriter, src)
	return nil, err
}

func (i GzipDecompressTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	decompressReader, err := gzip.NewReader(src)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(dst, decompressReader)
	return nil, err
}