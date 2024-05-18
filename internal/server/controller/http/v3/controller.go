package v3

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/arxon31/metrics-collector/internal/entity"
	repo "github.com/arxon31/metrics-collector/internal/repository"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const (
	saveJSONMetricsURL = "/updates/"
	getJSONMetricURL   = "/value/"
	getJSONMetricsURL  = "/"
)

type storageService interface {
	SaveBatchMetrics(ctx context.Context, metrics []entity.MetricDTO) error
}

type providerService interface {
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetMetrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type v3 struct {
	store    storageService
	provider providerService
}

func NewController(store storageService, provider providerService) *v3 {
	return &v3{
		store:    store,
		provider: provider,
	}
}

func (v *v3) Register(h *chi.Mux) {
	h.Post(saveJSONMetricsURL, v.saveJSONMetrics)
	h.Post(getJSONMetricURL, v.getJSONMetric)
	h.Get(getJSONMetricsURL, v.getJSONMetrics)
}

func (v *v3) saveJSONMetrics(w http.ResponseWriter, r *http.Request) {

	ms := make([]entity.MetricDTO, 0)

	if err := json.NewDecoder(r.Body).Decode(&ms); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := v.store.SaveBatchMetrics(r.Context(), ms)
	if err != nil {
		http.Error(w, "failed to store batch of metrics", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (v *v3) getJSONMetric(w http.ResponseWriter, r *http.Request) {
	m := entity.MetricDTO{}

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	switch m.MetricType {
	case entity.GaugeType:
		val, err := v.provider.GetGaugeValue(r.Context(), m.Name)
		if err != nil {
			if errors.Is(err, repo.ErrMetricNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m.Gauge = &val
	case entity.CounterType:
		val, err := v.provider.GetCounterValue(r.Context(), m.Name)
		if err != nil {
			if errors.Is(err, repo.ErrMetricNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m.Counter = &val
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

func (v *v3) getJSONMetrics(w http.ResponseWriter, r *http.Request) {
	ms, err := v.provider.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(ms)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
