package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/arxon31/metrics-collector/pkg/logger"
)

const hashHeader = "HashSHA256"

type hashingMiddleware struct {
	key string
}

func NewHashingMiddleware(key string) *hashingMiddleware {
	return &hashingMiddleware{
		key: key,
	}
}

type hashingResponseWriter struct {
	http.ResponseWriter
	key string
}

// WithHash middleware adds sha256 hash to request body if key is not empty
func (h *hashingMiddleware) WithHash(next http.Handler) http.Handler {
	if h.key == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(hashHeader) == "" {
			logger.Logger.Debug("body after no hashing: ", r.Body)
			next.ServeHTTP(w, r)
			return
		}

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Logger.Error(err)
			http.Error(w, "can not read body", http.StatusInternalServerError)
			return
		}

		sign, err := countHash(buf, h.key)
		if err != nil {
			logger.Logger.Error(err)
			http.Error(w, "can not count hash for body", http.StatusInternalServerError)
			return
		}
		signFromReq := r.Header.Get(hashHeader)
		if sign != signFromReq {
			logger.Logger.Error("signs is not equal")
			http.Error(w, "signs is not equal", http.StatusBadRequest)
			return
		}

		hw := &hashingResponseWriter{
			ResponseWriter: w,
			key:            h.key,
		}

		r.Body = io.NopCloser(bytes.NewReader(buf))

		next.ServeHTTP(hw, r)
	})
}

func (w *hashingResponseWriter) Write(data []byte) (size int, err error) {
	hash, err := countHash(data, w.key)
	if err != nil {
		return 0, err
	}

	w.ResponseWriter.Header().Add("HashSHA256", hash)

	return w.ResponseWriter.Write(data)
}

func countHash(data []byte, key string) (hash string, err error) {
	var sign []byte

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	sign = h.Sum(nil)

	return base64.StdEncoding.EncodeToString(sign), nil
}
