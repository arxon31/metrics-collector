package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"go.uber.org/zap"
	"net/http/httptest"
	"testing"
)

// BenchmarkPostJSONBatch benchmarks the ServeHTTP for PostJSONBatch handler
func BenchmarkPostJSONBatch(b *testing.B) {
	b.StopTimer()
	metricCount := 100
	repo := memory.NewMapStorage()
	logger, _ := zap.NewDevelopment()
	sugared := logger.Sugar()
	postBatchJSON := &PostJSONBatch{Storage: repo, Provider: repo, Logger: sugared}
	exampleGauge := 10.0
	exampleCounter := int64(10)
	metrics := []metric.MetricDTO{}
	for i := 0; i < metricCount; i++ {
		metrics = append(metrics, metric.MetricDTO{
			ID:    fmt.Sprintf("metric%d", i),
			MType: "gauge",
			Value: &exampleGauge,
		})
		metrics = append(metrics, metric.MetricDTO{
			ID:    fmt.Sprintf("metric%d", i),
			MType: "counter",
			Delta: &exampleCounter,
		})
	}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		reqBody, err := json.Marshal(metrics)
		if err != nil {
			b.Error(err)
		}
		body := bytes.NewBuffer(reqBody)
		req := httptest.NewRequest("POST", "/update", body)
		req.Header.Set("Content-Type", "application/json")
		b.StartTimer()
		postBatchJSON.ServeHTTP(w, req)
	}
}
