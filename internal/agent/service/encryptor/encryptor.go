package encryptor

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
)

type encryptor struct {
	key *rsa.PublicKey
}

// NewEncryptorService creates new encryptor service
func NewEncryptorService(key *rsa.PublicKey) *encryptor {
	return &encryptor{
		key: key,
	}
}

// Encrypt encrypts data by provided key
func (e *encryptor) Encrypt(data []byte) ([]byte, error) {
	var encryptedData bytes.Buffer

	chunkSize := e.key.Size() - 11

	chunks := len(data) / chunkSize

	for i := 0; i <= chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}
		encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, e.key, data[start:end])
		if err != nil {
			return nil, err
		}
		_, err = encryptedData.Write(encryptedChunk)
		if err != nil {
			return nil, err
		}
	}

	return encryptedData.Bytes(), nil
}
