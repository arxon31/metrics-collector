package middlewares

import (
	"compress/gzip"
	"github.com/arxon31/metrics-collector/pkg/e"
	"io"
	"log"
	"net/http"
	"strings"
)

var compressibleTypes = map[string]bool{
	"text/html":        true,
	"application/json": true,
}

type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w compressWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func (r compressReader) Read(b []byte) (int, error) {
	return r.zr.Read(b)
}

func (r compressReader) Close() error {
	if err := r.Close(); err != nil {
		return err
	}

	return r.zr.Close()
}

func WithCompressing(next http.Handler) http.Handler {
	const op = "middlewares.WithCompressing()"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			//w.Header().Set("Content-Encoding", "gzip")

			writer = compressWriter{
				ResponseWriter: w,
				Writer:         gzipWriter,
			}

		}
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Println(e.WrapString(op, "failed to create gzip reader", err))
				http.Error(w, e.WrapString(op, "failed to create gzip reader", err), http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()

			//reader := compressReader{
			//	ReadCloser: r.Body,
			//	zr:         gzipReader,
			//}

			r.Body = gzipReader

		}

		next.ServeHTTP(writer, r)
	})
}
