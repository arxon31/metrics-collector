package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	repo "github.com/arxon31/metrics-collector/internal/repository/repoerr"
)

func TestV1_NewController(t *testing.T) {
	store := &storageServiceMock{}
	provider := &providerServiceMock{}
	v := NewController(store, provider)
	require.IsType(t, &v1{}, v)
}

func TestV1_GetGaugeMetric(t *testing.T) {

	t.Run("get_gauge_metric_success", func(t *testing.T) {
		provider := &providerServiceMock{
			GetGaugeValueFunc: func(ctx context.Context, name string) (float64, error) {
				return 20.1, nil
			},
		}

		v := &v1{
			provider: provider,
		}
		req := httptest.NewRequest(http.MethodGet, getGaugeMetricURL, nil)
		rr := httptest.NewRecorder()
		v.getGaugeMetric(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "20.1", rr.Body.String())
		require.Equal(t, "text/html; charset=utf-8", rr.Header().Get("Content-Type"))
	})
	t.Run("get_gauge_metric_not_found", func(t *testing.T) {
		provider := &providerServiceMock{
			GetGaugeValueFunc: func(ctx context.Context, name string) (float64, error) {
				return 0, repo.ErrMetricNotFound
			},
		}
		v := &v1{
			provider: provider,
		}
		req := httptest.NewRequest(http.MethodGet, getGaugeMetricURL, nil)
		rr := httptest.NewRecorder()
		v.getGaugeMetric(rr, req)
		require.Equal(t, http.StatusNotFound, rr.Code)

	})

}

func TestV1_GetCounterMetric(t *testing.T) {
	t.Run("get_counter_metric_success", func(t *testing.T) {
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return 20, nil
			},
		}
		v := &v1{
			provider: provider,
		}
		req := httptest.NewRequest(http.MethodGet, getCounterMetricURL, nil)
		rr := httptest.NewRecorder()
		v.getCounterMetric(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "20", rr.Body.String())
		require.Equal(t, "text/html; charset=utf-8", rr.Header().Get("Content-Type"))

	})
	t.Run("get_counter_metric_not_found", func(t *testing.T) {
		provider := &providerServiceMock{
			GetCounterValueFunc: func(ctx context.Context, name string) (int64, error) {
				return 0, repo.ErrMetricNotFound
			},
		}
		v := &v1{
			provider: provider,
		}
		req := httptest.NewRequest(http.MethodGet, getCounterMetricURL, nil)
		rr := httptest.NewRecorder()
		v.getCounterMetric(rr, req)
		require.Equal(t, http.StatusNotFound, rr.Code)
	})
}
