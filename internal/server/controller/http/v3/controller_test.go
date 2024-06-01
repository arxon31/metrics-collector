package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arxon31/metrics-collector/internal/entity"
)

func TestV3_NewController(t *testing.T) {
	store := &storageServiceMock{}
	provider := &providerServiceMock{}
	pinger := &pingerServiceMock{}
	v := NewController(store, provider, pinger)
	require.IsType(t, &v3{}, v)
}

func TestV3_PingDB(t *testing.T) {
	t.Run("ping_db_success", func(t *testing.T) {
		pinger := &pingerServiceMock{
			PingDBFunc: func() error {
				return nil
			},
		}
		v3 := NewController(&storageServiceMock{}, &providerServiceMock{}, pinger)
		req := httptest.NewRequest(http.MethodGet, pingDBURL, nil)
		rr := httptest.NewRecorder()
		v3.pingDB(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("ping_db_fail", func(t *testing.T) {
		pinger := &pingerServiceMock{
			PingDBFunc: func() error {
				return errors.New("some error")
			},
		}
		v3 := NewController(&storageServiceMock{}, &providerServiceMock{}, pinger)
		req := httptest.NewRequest(http.MethodGet, pingDBURL, nil)
		rr := httptest.NewRecorder()
		v3.pingDB(rr, req)
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestV3_SaveJSONMetrics(t *testing.T) {
	t.Run("save_json_metrics_success", func(t *testing.T) {

		store := &storageServiceMock{
			SaveBatchMetricsFunc: func(ctx context.Context, metrics []entity.MetricDTO) error {
				return nil
			},
		}

		v3 := NewController(store, &providerServiceMock{}, &pingerServiceMock{})

		gaugeVal := 20.1
		counterVal := int64(20)
		metrics := []entity.MetricDTO{
			{
				Name:       "test",
				MetricType: "gauge",
				Gauge:      &gaugeVal,
			},
			{
				Name:       "test",
				MetricType: "counter",
				Counter:    &counterVal,
			},
		}

		metricsJSON, err := json.Marshal(metrics)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, saveJSONMetricsURL, bytes.NewBuffer(metricsJSON))
		rr := httptest.NewRecorder()
		v3.saveJSONMetrics(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("save_json_metrics_fail", func(t *testing.T) {
		store := &storageServiceMock{
			SaveBatchMetricsFunc: func(ctx context.Context, metrics []entity.MetricDTO) error {
				return errors.New("some error")
			},
		}

		v3 := NewController(store, &providerServiceMock{}, &pingerServiceMock{})

		gaugeVal := 20.1
		counterVal := int64(20)
		metrics := []entity.MetricDTO{
			{
				Name:       "test",
				MetricType: "gauge",
				Gauge:      &gaugeVal,
			},
			{
				Name:       "test",
				MetricType: "counter",
				Counter:    &counterVal,
			},
		}

		metricsJSON, err := json.Marshal(metrics)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, saveJSONMetricsURL, bytes.NewBuffer(metricsJSON))
		rr := httptest.NewRecorder()
		v3.saveJSONMetrics(rr, req)
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("save_json_metrics_bad_request", func(t *testing.T) {
		v3 := NewController(&storageServiceMock{}, &providerServiceMock{}, &pingerServiceMock{})
		req := httptest.NewRequest(http.MethodPost, saveJSONMetricsURL, nil)
		rr := httptest.NewRecorder()
		v3.saveJSONMetrics(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestV3_GetJSONMetrics(t *testing.T) {
	t.Run("get_json_metrics_success", func(t *testing.T) {
		gaugeVal := 20.1
		counterVal := int64(20)

		metrics := []entity.MetricDTO{
			{
				Name:       "test",
				MetricType: "gauge",
				Gauge:      &gaugeVal,
			},
			{
				Name:       "test",
				MetricType: "counter",
				Counter:    &counterVal,
			},
		}
		provider := &providerServiceMock{
			GetMetricsFunc: func(ctx context.Context) ([]entity.MetricDTO, error) {
				return metrics, nil
			},
		}

		metricsJSON, err := json.Marshal(metrics)
		require.NoError(t, err)

		v3 := NewController(&storageServiceMock{}, provider, &pingerServiceMock{})
		req := httptest.NewRequest(http.MethodGet, getJSONMetricsURL, nil)
		rr := httptest.NewRecorder()
		v3.getJSONMetrics(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, string(metricsJSON), rr.Body.String())
	})

	t.Run("get_json_metrics_fail", func(t *testing.T) {
		provider := &providerServiceMock{
			GetMetricsFunc: func(ctx context.Context) ([]entity.MetricDTO, error) {
				return nil, errors.New("some error")
			},
		}
		v3 := NewController(&storageServiceMock{}, provider, &pingerServiceMock{})
		req := httptest.NewRequest(http.MethodGet, getJSONMetricsURL, nil)
		rr := httptest.NewRecorder()
		v3.getJSONMetrics(rr, req)
		require.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
