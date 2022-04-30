package transforms

import (
	"compress/gzip"
	"io"
)

type GzipCompressTransform struct {
	args map[string]interface{}
}

type GzipDecompressTransform struct {
}

func (i GzipCompressTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	var compressWriter io.WriteCloser
	var err error
	if _, ok := i.args["level"]; ok {
		compressWriter, err = gzip.NewWriterLevel(dst, i.args["level"].(int))
		if err != nil {
			return nil, err
		}
	} else {
		compressWriter = gzip.NewWriter(dst)
	}
	_, err = io.Copy(compressWriter, src)
	compressWriter.Close()
	return nil, err
}

func (i GzipDecompressTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	decompressReader, err := gzip.NewReader(src)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(dst, decompressReader)
	decompressReader.Close()
	return nil, err
}