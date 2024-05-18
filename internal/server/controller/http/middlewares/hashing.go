package middlewares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
)

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
		var data []byte
		_, err := r.Body.Read(data)
		if err != nil {
			http.Error(w, "can not read body", http.StatusInternalServerError)
		}
		sign, err := countHash(data, h.key)
		if err != nil {
			http.Error(w, "can not count hash for body", http.StatusInternalServerError)
			return
		}
		signFromReq := r.Header.Get("HashSHA256")
		if sign != signFromReq {
			http.Error(w, "signs is not equal", http.StatusBadRequest)
			return
		}

		hw := &hashingResponseWriter{
			ResponseWriter: w,
			key:            h.key,
		}
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
