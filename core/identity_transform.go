package core

type IdentityTransform struct {
}

func (i IdentityTransform) transform(src []byte) []byte {
	return src
}

