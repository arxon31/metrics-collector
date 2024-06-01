package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/arxon31/metrics-collector/internal/repository/repoerr"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestV2_NewController(t *testing.T) {
	store := &storageServiceMock{}
	provider := &providerServiceMock{}
	v := NewController(store, provider)
	require.IsType(t, &v2{}, v)
}

func TestV2_UpdateJSONMetric(t *testing.T) {
	t.Run("update_json_gauge_metric_success", func(t *testing.T) {
		store := &storageServiceMock{
			SaveGaugeMetricFunc: func(ctx context.Context, m entity.MetricDTO) error {
				return nil
			},
		}
		provider := &providerServiceMock{}
		v2 := NewController(store, provider)

		gaugeVal := 20.1
		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "gauge",
			Gauge:      &gaugeVal,
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewReader(metricJSON))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))
		require.Equal(t, string(metricJSON), w.Body.String())
	})
	t.Run("update_json_counter_metric_success", func(t *testing.T) {
		counterVal := int64(20)

		store := &storageServiceMock{
			SaveCounterMetricFunc: func(ctx context.Context, m entity.MetricDTO) error {
				return nil
			},
		}
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return counterVal, nil
			},
		}
		v2 := NewController(store, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "counter",
			Counter:    &counterVal,
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewReader(metricJSON))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "application/json", w.Header().Get("Content-Type"))
		require.Equal(t, string(metricJSON), w.Body.String())
	})

	t.Run("update_json_metric_bad_request", func(t *testing.T) {
		store := &storageServiceMock{}
		provider := &providerServiceMock{}
		v2 := NewController(store, provider)

		metric := entity.MetricDTO{
			MetricType: "unknown",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update_gauge_metric_save_error", func(t *testing.T) {
		store := &storageServiceMock{
			SaveGaugeMetricFunc: func(ctx context.Context, m entity.MetricDTO) error {
				return errors.New("some error")
			},
		}
		provider := &providerServiceMock{}
		v2 := NewController(store, provider)

		gaugeVal := 20.1
		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "gauge",
			Gauge:      &gaugeVal,
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update_counter_metric_save_error", func(t *testing.T) {
		store := &storageServiceMock{
			SaveCounterMetricFunc: func(ctx context.Context, m entity.MetricDTO) error {
				return errors.New("some error")
			},
		}
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return -1, repoerr.ErrMetricNotFound
			},
		}
		v2 := NewController(store, provider)

		counterVal := int64(20)
		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "counter",
			Counter:    &counterVal,
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("update_counter_metric_not_found", func(t *testing.T) {
		store := &storageServiceMock{
			SaveCounterMetricFunc: func(ctx context.Context, m entity.MetricDTO) error {
				return nil
			},
		}
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return -1, repoerr.ErrMetricNotFound
			},
		}
		v2 := NewController(store, provider)

		counterVal := int64(20)
		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "counter",
			Counter:    &counterVal,
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error_can_not_decode_metric", func(t *testing.T) {
		store := &storageServiceMock{}
		provider := &providerServiceMock{}
		v2 := NewController(store, provider)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewBuffer([]byte("")))
		w := httptest.NewRecorder()
		v2.updateJSONMetric(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestV2_GetValueOfJSONMetric(t *testing.T) {
	t.Run("get_value_of_json_metric_gauge_success", func(t *testing.T) {
		gaugeVal := 20.1
		provider := &providerServiceMock{
			GetGaugeValueFunc: func(ctx context.Context, name string) (float64, error) {
				return gaugeVal, nil
			},
		}
		v2 := NewController(&storageServiceMock{}, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "gauge",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		respMetric := metric
		respMetric.Gauge = &gaugeVal
		respMetricJSON, err := json.Marshal(respMetric)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, string(respMetricJSON), w.Body.String())
	})

	t.Run("get_value_of_json_metric_counter_success", func(t *testing.T) {
		counterVal := int64(20)
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return counterVal, nil
			},
		}
		v2 := NewController(&storageServiceMock{}, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "counter",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		respMetric := metric
		respMetric.Counter = &counterVal
		respMetricJSON, err := json.Marshal(respMetric)

		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, string(respMetricJSON), w.Body.String())
	})

	t.Run("get_value_of_json_metric_gauge_not_found", func(t *testing.T) {
		provider := &providerServiceMock{
			GetGaugeValueFunc: func(ctx context.Context, name string) (float64, error) {
				return -1, repoerr.ErrMetricNotFound
			},
		}
		v2 := NewController(&storageServiceMock{}, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "gauge",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get_value_of_json_metric_counter_not_found", func(t *testing.T) {
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return -1, repoerr.ErrMetricNotFound
			},
		}
		v2 := NewController(&storageServiceMock{}, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "counter",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get_value_of_json_metric_gauge_other_repo_error", func(t *testing.T) {
		provider := &providerServiceMock{
			GetGaugeValueFunc: func(ctx context.Context, name string) (float64, error) {
				return -1, errors.New("some error")
			},
		}
		v2 := NewController(&storageServiceMock{}, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "gauge",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("get_value_of_json_metric_counter_other_repo_error", func(t *testing.T) {
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return -1, errors.New("some error")
			},
		}
		v2 := NewController(&storageServiceMock{}, provider)

		metric := entity.MetricDTO{
			Name:       "test",
			MetricType: "counter",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		require.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("error_can_not_decode_metric", func(t *testing.T) {
		store := &storageServiceMock{}
		provider := &providerServiceMock{}
		v2 := NewController(store, provider)

		r := httptest.NewRequest(http.MethodPost, valueOfMetricJSONURL, bytes.NewBuffer([]byte("")))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)

		require.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("update_json_metric_bad_request", func(t *testing.T) {
		store := &storageServiceMock{}
		provider := &providerServiceMock{}
		v2 := NewController(store, provider)

		metric := entity.MetricDTO{
			MetricType: "unknown",
		}

		metricJSON, err := json.Marshal(metric)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, updateMetricJSONURL, bytes.NewBuffer(metricJSON))
		w := httptest.NewRecorder()
		v2.getValueOfJSONMetric(w, r)
		require.Equal(t, http.StatusBadRequest, w.Code)
	})

}
