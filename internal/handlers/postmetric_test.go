package handlers

import (
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostMetrics(t *testing.T) {
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
			name:     "Successful request[gauge]",
			endpoint: "http://localhost:8080/update/gauge/testMetric/10.0001",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:     "Without metric name[gauge]",
			endpoint: "http://localhost:8080/update/gauge/10.0001",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "With unsupported metric value[gauge]",
			endpoint: "http://localhost:8080/update/gauge/testMetric/value",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "Successful request [counter]",
			endpoint: "http://localhost:8080/update/counter/testMetric/10",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:     "Without metric name [counter]",
			endpoint: "http://localhost:8080/update/counter/10",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		}, {
			name:     "With unsupported metric value [counter]",
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

			var postMetrics PostMetrics
			postMetrics.Storage = mem.NewMapStorage()

			router := chi.NewRouter()
			router.Post("/update/{type}/{name}/{value}", postMetrics.ServeHTTP)
			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
