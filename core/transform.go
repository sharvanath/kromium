package core

type Transform interface {
	// The Transformation to apply on source file when read as byte string.
	transform([]byte) []byte
}

func getTransform(name string) Transform {
	if name == "Identity" {
		return IdentityTransform{}
	}
	return nil
}
