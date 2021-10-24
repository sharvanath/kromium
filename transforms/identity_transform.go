package transforms

import (
	"io"
)

type IdentityTransform struct {
}

func (i IdentityTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	_, err := io.Copy(dst, src)
	return nil, err
}

