package handlers

import (
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGaugeHandler(t *testing.T) {
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
			endpoint: "http://localhost:8080/update/gauge/testMetric/10.0001",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:     "Without metric name",
			endpoint: "http://localhost:8080/update/gauge/10.0001",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "Without metric value",
			endpoint: "http://localhost:8080/update/gauge/testMetric/",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "With unsupported metric value",
			endpoint: "http://localhost:8080/update/gauge/testMetric/value",
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

			var gauge GaugeHandler
			gauge.Storage = mem.NewMapStorage()
			gauge.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
