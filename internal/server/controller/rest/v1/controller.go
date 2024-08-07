package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/arxon31/metrics-collector/internal/server/controller/rest/resterrs"

	"github.com/go-chi/chi/v5"

	"github.com/arxon31/metrics-collector/internal/entity"
	repo "github.com/arxon31/metrics-collector/internal/repository/repoerr"
)

const (
	saveCounterMetricURL = "/counter/{name}/{value}"
	saveGaugeMetricURL   = "/gauge/{name}/{value}"
	saveUnimplementedURL = "/{type}/{name}/{value}"
	getGaugeMetricURL    = "/value/gauge/{name}"
	getCounterMetricURL  = "/value/counter/{name}"
	getUnimplementedURL  = "/value/{type}/{name}"
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
}

type v1 struct {
	store    storageService
	provider providerService
}

// NewController initializes a new v3 controller.
func NewController(store storageService, provider providerService) *v1 {
	return &v1{
		store:    store,
		provider: provider,
	}
}

// Register registers the v1 endpoints on the provided chi Router.
func (v *v1) Register(h *chi.Mux) {
	h.Route("/value", func(r chi.Router) {
		h.Get(getGaugeMetricURL, v.getGaugeMetric)
		h.Get(getCounterMetricURL, v.getCounterMetric)
		h.Get(getUnimplementedURL, v.unimplementedGet)
	})
	h.Route("/update", func(r chi.Router) {
		r.Post(saveCounterMetricURL, v.updateCounterMetric)
		r.Post(saveGaugeMetricURL, v.updateGaugeMetric)
		r.Post(saveUnimplementedURL, v.unimplementedSave)
	})
}

func (v *v1) getGaugeMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value, err := v.provider.GetGaugeValue(r.Context(), name)
	if err != nil {
		if errors.Is(err, repo.ErrMetricNotFound) {
			http.Error(w, fmt.Sprintf("%s: %s", resterrs.ErrMetricNotFound, name), http.StatusNotFound)
			return
		}
		http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v", value)))

}

func (v *v1) getCounterMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	value, err := v.provider.GetCounterValue(r.Context(), name)
	if err != nil {
		if errors.Is(err, repo.ErrMetricNotFound) {
			http.Error(w, fmt.Sprintf("%s: %s", resterrs.ErrMetricNotFound, name), http.StatusNotFound)
			return
		}
		http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v", value)))
}

func (v *v1) updateCounterMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	fmt.Println(name, value)

	var counter entity.Counter
	val, err := counter.CounterFromString(value)
	if err != nil {
		http.Error(w, resterrs.ErrUnexpectedValue.Error(), http.StatusBadRequest)
		return
	}

	err = v.store.SaveCounterMetric(r.Context(), entity.MetricDTO{
		MetricType: entity.CounterType,
		Name:       name,
		Counter:    &val,
	})

	if err != nil {
		http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (v *v1) updateGaugeMetric(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var gauge entity.Gauge
	val, err := gauge.GaugeFromString(value)
	if err != nil {
		http.Error(w, resterrs.ErrUnexpectedValue.Error(), http.StatusBadRequest)
		return
	}

	err = v.store.SaveGaugeMetric(r.Context(), entity.MetricDTO{
		MetricType: entity.GaugeType,
		Name:       name,
		Gauge:      &val,
	})

	if err != nil {
		http.Error(w, resterrs.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (v *v1) unimplementedSave(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	if t != entity.GaugeType && t != entity.CounterType {
		http.Error(w, resterrs.ErrUnexpectedType.Error(), http.StatusNotImplemented)
		return
	}
}

func (v *v1) unimplementedGet(w http.ResponseWriter, r *http.Request) {
	t := chi.URLParam(r, "type")
	if t != entity.GaugeType && t != entity.CounterType {
		http.Error(w, resterrs.ErrUnexpectedType.Error(), http.StatusNotImplemented)
		return
	}
}
