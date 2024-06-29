package v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/arxon31/metrics-collector/internal/server/controller/rest/resterrs"

	repo "github.com/arxon31/metrics-collector/internal/repository/repoerr"

	"github.com/go-chi/chi/v5"

	"github.com/arxon31/metrics-collector/internal/entity"
)

const (
	updateMetricJSONURL  = "/update/"
	valueOfMetricJSONURL = "/value/"
)

//go:generate moq -out storageService_moq_test.go . storageService
type storageService interface {
	SaveGaugeMetric(ctx context.Context, metric entity.MetricDTO) error
	SaveCounterMetric(ctx context.Context, metric entity.MetricDTO) error
}

//go:generate moq -out providerService_moq_test.go . providerService
type providerService interface {
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetMetrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type v2 struct {
	store    storageService
	provider providerService
}

// NewController initializes a new v3 controller.
func NewController(store storageService, provider providerService) *v2 {
	return &v2{
		store:    store,
		provider: provider,
	}
}

// Register registers the v2 endpoints on the provided chi Router.
func (v *v2) Register(h *chi.Mux) {
	h.Post(updateMetricJSONURL, v.updateJSONMetric)
	h.Post(valueOfMetricJSONURL, v.getValueOfJSONMetric)

}

func (v *v2) updateJSONMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m entity.MetricDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, resterrs.ErrUnexpectedFormat.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if m.MetricType != entity.GaugeType && m.MetricType != entity.CounterType {
		http.Error(w, resterrs.ErrUnexpectedType.Error(), http.StatusBadRequest)
		return
	}

	switch m.MetricType {
	case entity.GaugeType:
		err := v.store.SaveGaugeMetric(r.Context(), m)
		if err != nil {
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

	case entity.CounterType:
		err := v.store.SaveCounterMetric(r.Context(), m)
		if err != nil {
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}

		counterValue, err := v.provider.GetCounterValue(r.Context(), m.Name)
		if err != nil {
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		}

		m.Counter = &counterValue

	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (v *v2) getValueOfJSONMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	m := entity.MetricDTO{}

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, resterrs.ErrUnexpectedFormat.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if m.MetricType != entity.GaugeType && m.MetricType != entity.CounterType {
		http.Error(w, resterrs.ErrUnexpectedType.Error(), http.StatusBadRequest)
		return
	}

	switch m.MetricType {
	case entity.GaugeType:
		val, err := v.provider.GetGaugeValue(r.Context(), m.Name)
		if err != nil {
			if errors.Is(err, repo.ErrMetricNotFound) {
				http.Error(w, fmt.Sprintf("%s: %s", resterrs.ErrMetricNotFound, m.Name), http.StatusNotFound)
				return
			}
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}
		m.Gauge = &val
	case entity.CounterType:
		val, err := v.provider.GetCounterValue(r.Context(), m.Name)
		if err != nil {
			if errors.Is(err, repo.ErrMetricNotFound) {
				http.Error(w, fmt.Sprintf("%s: %s", resterrs.ErrMetricNotFound, m.Name), http.StatusNotFound)
				return
			}
			http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
			return
		}
		m.Counter = &val
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
