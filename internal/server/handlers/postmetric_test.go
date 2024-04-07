package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/arxon31/metrics-collector/internal/repository/memory"
)

func TestPostGaugeMetric(t *testing.T) {
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

			st := memory.NewMapStorage()

			var postGaugeMetric PostGaugeMetric
			postGaugeMetric.Storage = st
			postGaugeMetric.Provider = st

			router := chi.NewRouter()
			router.Post("/update/gauge/{name}/{value}", postGaugeMetric.ServeHTTP)
			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}

func TestPostCounterMetric(t *testing.T) {
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

			st := memory.NewMapStorage()

			var postCounterMetric PostCounterMetrics
			postCounterMetric.Storage = st
			postCounterMetric.Provider = st

			router := chi.NewRouter()
			router.Post("/update/counter/{name}/{value}", postCounterMetric.ServeHTTP)
			router.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
