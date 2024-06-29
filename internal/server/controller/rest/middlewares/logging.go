package middlewares

import (
	"net/http"
	"time"

	"github.com/arxon31/metrics-collector/pkg/logger"
)

type loggingMiddleware struct {
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

type responseData struct {
	statusCode int
	size       int
}

func NewLoggingMiddleware() *loggingMiddleware {
	return &loggingMiddleware{}
}

// WithLog middleware logs requests
func (l *loggingMiddleware) WithLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		respData := &responseData{}
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   respData,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		logger.Logger.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"execution_time", duration,
			"status_code", respData.statusCode,
			"size", respData.size,
		)

	})
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.statusCode = statusCode
}
