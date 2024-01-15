package middlewares

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

type responseData struct {
	statusCode int
	size       int
}

func WithLogging(logger *zap.SugaredLogger, next http.Handler) http.Handler {
	const op = "middlewares.WithLogging()"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		respData := &responseData{
			statusCode: 0,
			size:       0,
		}
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   respData,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		logger.Infoln(
			"function", op,
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
