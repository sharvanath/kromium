package transforms

import (
	"io"
	"strings"
	"github.com/rwtodd/Go.Sed/sed"
)

type SedTransform struct {
	arg string
}

func (s SedTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	engine, err := sed.New(strings.NewReader(s.arg))
	_, err = io.Copy(dst, engine.Wrap(src))
	return nil, err
}
