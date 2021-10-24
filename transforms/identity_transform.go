package transforms

import (
	"fmt"
	"io"
)

type IdentityTransform struct {
}

func (i IdentityTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	num, err := io.Copy(dst, src)
	fmt.Printf("Copied %d bytes\n", num)
	return nil, err
}

