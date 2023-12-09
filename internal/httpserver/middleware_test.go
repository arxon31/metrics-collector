package httpserver

import (
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		method         string
		wantStatusCode int
	}{
		{
			name:           "POST request",
			endpoint:       "http://localhost:8080/update/gauge/testMetric/10.0001",
			method:         http.MethodPost,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "GET request",
			endpoint:       "http://localhost:8080/update/gauge/testMetric/10.0001",
			method:         http.MethodGet,
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()
			storage := mem.NewMapStorage()
			handler := &handlers.PostMetrics{Storage: storage}
			mw := postCheck(handler)
			mw.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
		})
	}
}
