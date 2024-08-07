package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var errNoHashKey = errors.New("no hash key provided")

type hasher struct {
	hashKey string
}

// NewHasherService creates new hasher service
func NewHasherService(hashKey string) *hasher {
	return &hasher{
		hashKey: hashKey,
	}
}

// Hash hashes data by provided hash key
func (h *hasher) Hash(data []byte) (sign string, err error) {
	if h.hashKey == "" {
		return "", errNoHashKey
	}
	hash := hmac.New(sha256.New, []byte(h.hashKey))
	hash.Write(data)
	s := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(s), nil
}
