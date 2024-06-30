package middlewares

import (
	"compress/gzip"
	"fmt"
	"github.com/arxon31/metrics-collector/pkg/logger"
	"io"
	"net/http"
	"strings"
)

var compressibleTypes = map[string]bool{
	"text/html":        true,
	"application/json": true,
	"html/text":        true,
}

type compressingMiddleware struct {
}

func NewCompressingMiddleware() *compressingMiddleware {
	return &compressingMiddleware{}
}

// WithCompress middleware compresses and decompresses responses
func (c *compressingMiddleware) WithCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && (compressibleTypes[r.Header.Get("Content-Type")] || compressibleTypes[r.Header.Get("Accept")]) {
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			writer = compressWriter{
				ResponseWriter: w,
				Writer:         gzipWriter,
			}

			w.Header().Set("Content-Encoding", "gzip")

		}
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Logger.Error(err)
				http.Error(w, fmt.Sprintf("can not create gzip reader: %s", err), http.StatusInternalServerError)
				return
			}
			defer gzipReader.Close()

			r.Body = io.NopCloser(gzipReader)

		}

		next.ServeHTTP(writer, r)
	})
}

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w compressWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
