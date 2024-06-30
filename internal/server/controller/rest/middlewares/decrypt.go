package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"github.com/arxon31/metrics-collector/internal/server/controller/rest/resterrs"
	"github.com/arxon31/metrics-collector/pkg/logger"
	"io"
	"net/http"
)

type decryptingMiddleware struct {
	cryptoKey *rsa.PrivateKey
}

func NewDecryptingMiddleware(cryptoKey *rsa.PrivateKey) *decryptingMiddleware {
	return &decryptingMiddleware{
		cryptoKey: cryptoKey,
	}
}

func (d *decryptingMiddleware) WithDecrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Logger.Error(err)
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		decryptedBody, err := d.decrypt(body)
		if err != nil {
			logger.Logger.Error(err)
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(decryptedBody))

		next.ServeHTTP(w, r)
	})
}

func (d *decryptingMiddleware) decrypt(body []byte) ([]byte, error) {
	chunkSize := d.cryptoKey.Size() - 11

	chunks := len(body) / chunkSize

	var decryptedData bytes.Buffer

	for i := 0; i <= chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(body) {
			end = len(body)
		}

		decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, d.cryptoKey, body[start:end])
		if err != nil {
			return nil, err
		}
		decryptedData.Write(decryptedChunk)
	}

	return decryptedData.Bytes(), nil
}
