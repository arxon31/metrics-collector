package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arxon31/metrics-collector/internal/entity"
)

func BenchmarkV3_SaveJSONMetrics(b *testing.B) {
	store := &storageServiceMock{
		SaveBatchMetricsFunc: func(ctx context.Context, metrics []entity.MetricDTO) error {
			return nil
		},
	}
	provider := &providerServiceMock{}
	pinger := &pingerServiceMock{}
	v3 := NewController(store, provider, pinger)

	testGauge := 20.1
	testCounter := int64(20)

	metrics := []entity.MetricDTO{}

	for i := 0; i < 1000; i++ {
		metrics = append(metrics, entity.MetricDTO{
			Name:       fmt.Sprintf("metric_counter_%d", i),
			MetricType: "counter",
			Counter:    &testCounter,
		})
		metrics = append(metrics, entity.MetricDTO{
			Name:       fmt.Sprintf("metric_gauge_%d", i),
			MetricType: "gauge",
			Gauge:      &testGauge,
		})
	}

	metricsJSON, err := json.Marshal(metrics)
	require.NoError(b, err)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, saveJSONMetricsURL, bytes.NewBuffer(metricsJSON))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		v3.saveJSONMetrics(w, req)
	}
}
