package middlewares

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/arxon31/metrics-collector/pkg/e"
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
	const op = "middlewares.WithCompress()"
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
				log.Println(e.WrapString(op, "failed to create gzip reader", err))
				http.Error(w, e.WrapString(op, "failed to create gzip reader", err), http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			r.Body = gzipReader

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
