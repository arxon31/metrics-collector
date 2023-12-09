package handlers

import (
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCounterHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name     string
		endpoint string
		want     want
	}{
		{
			name:     "Successful request",
			endpoint: "http://localhost:8080/update/counter/testMetric/10",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:     "Without metric name",
			endpoint: "http://localhost:8080/update/counter/10",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "Without metric value",
			endpoint: "http://localhost:8080/update/counter/testMetric/",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "With unsupported metric value",
			endpoint: "http://localhost:8080/update/counter/testMetric/value",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.endpoint, nil)
			w := httptest.NewRecorder()

			var counter CounterHandler
			counter.Storage = mem.NewMapStorage()
			counter.ServeHTTP(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
