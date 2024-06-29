package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/arxon31/metrics-collector/internal/entity"
)

var (
	store = &storageServiceMock{
		SaveBatchMetricsFunc: func(ctx context.Context, metrics []entity.MetricDTO) error {
			return nil
		},
	}
	provider = &providerServiceMock{}
	pinger   = &pingerServiceMock{}

	counterTest = int64(20)
	gaugeTest   = float64(20.1)
)

func Example_v3saveJSONMetrics() {
	v3 := NewController(store, provider, pinger)

	metrics := []entity.MetricDTO{
		{
			Name:       "metric_counter_1",
			MetricType: "counter",
			Counter:    &counterTest,
		},
		{
			Name:       "metric_counter_2",
			MetricType: "counter",
			Counter:    &counterTest,
		},
		{
			Name:       "metric_gauge_1",
			MetricType: "gauge",
			Gauge:      &gaugeTest,
		},
		{
			Name:       "metric_gauge_2",
			MetricType: "gauge",
			Gauge:      &gaugeTest,
		},
	}

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return
	}

	req := httptest.NewRequest(http.MethodPost, saveJSONMetricsURL, bytes.NewBuffer(metricsJSON))
	rr := httptest.NewRecorder()
	v3.getJSONMetrics(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)

	// Output:
	// 200

}
