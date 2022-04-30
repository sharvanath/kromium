package transforms

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

type EncryptionTransform struct {
	HexKey string
}

type DecryptionTransform struct {
	HexKey string
}

func parseArgs(args map[string]interface{}, out interface{}) error {
	jsonStr, err := json.Marshal(args)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonStr, &out); err != nil {
		return err
	}
	return nil
}

func NewEncryptionTransform(args map[string]interface{}) (*EncryptionTransform) {
	var e EncryptionTransform
	if err := parseArgs(args, &e); err != nil {
		panic(err)
	}

	fmt.Printf("Key = %s\n", e.HexKey)

	return &e
}

func NewDecryptionTransform(args map[string]interface{}) (*DecryptionTransform) {
	var d DecryptionTransform
	if err := parseArgs(args, &d); err != nil {
		panic(err)
	}

	return &d
}

func (e EncryptionTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	key, _ := hex.DecodeString(e.HexKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	writer := &cipher.StreamWriter{S: stream, W: dst}
	// Copy the input to the output buffer, encrypting as we go.
	if _, err := io.Copy(writer, src); err != nil {
		panic(err)
	}
	writer.Close()
	return nil, err
}

func (e DecryptionTransform) Transform(dst io.Writer, src io.Reader) (interface{}, error) {
	key, _ := hex.DecodeString(e.HexKey)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	reader := &cipher.StreamReader{S: stream, R: src}
	// Copy the input to the output buffer, encrypting as we go.
	if _, err := io.Copy(dst, reader); err != nil {
		panic(err)
	}
	return nil, err
}