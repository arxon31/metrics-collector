package v2

import (
	"context"
	"encoding/json"
	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const (
	saveJSONMetricURL = "/update/"
)

type storageService interface {
	SaveGaugeMetric(ctx context.Context, metric entity.MetricDTO) error
	SaveCounterMetric(ctx context.Context, metric entity.MetricDTO) error
}

type v2 struct {
	store storageService
}

func NewController(store storageService) *v2 {
	return &v2{
		store: store,
	}
}

func (v *v2) Register(h *chi.Mux) {
	h.Post(saveJSONMetricURL, v.saveJSONMetric)
}

func (v *v2) saveJSONMetric(w http.ResponseWriter, r *http.Request) {
	var m entity.MetricDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if m.MetricType != entity.GaugeType && m.MetricType != entity.CounterType {
		http.Error(w, "wrong metric type", http.StatusBadRequest)
		return
	}
	switch m.MetricType {
	case entity.GaugeType:
		err := v.store.SaveGaugeMetric(r.Context(), m)
		if err != nil {
			http.Error(w, "can not save metric", http.StatusInternalServerError)
			return
		}
	case entity.CounterType:
		err := v.store.SaveCounterMetric(r.Context(), m)
		if err != nil {
			http.Error(w, "can not save metric", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
