package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
)

//type MetricCollector interface {
//	Replace(ctx context.Context, name string, value float64) error
//	Count(ctx context.Context, name string, value int64) error
//}

type Parser interface {
	GaugeFromString(value string) (float64, error)
	CounterFromString(value string) (int64, error)
}

type PostCounterMetrics Handler
type PostGaugeMetric Handler
type NotImplementedHandler Handler

// PostCounterMetrics implements http.Handler which gets the counter metric from URL and stores it in repository
func (h *PostCounterMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostCounterMetric.ServeHTTP()"

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var counter metric.Counter
	val, err := counter.CounterFromString(value)
	if err != nil {
		errStr := fmt.Sprintf("%v", e.WrapError(op, "value is invalid", err))
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	if err := h.Storage.Count(r.Context(), name, val); err != nil {
		errStr := fmt.Sprintf("%v", e.WrapError(op, "failed to replace metric", err))
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

// PostGaugeMetric implements http.Handler which gets the gauge metric from URL and stores it in repository
func (h *PostGaugeMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostGaugeMetric.ServeHTTP()"

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var gauge metric.Gauge

	val, err := gauge.GaugeFromString(value)
	if err != nil {
		errStr := fmt.Sprintf("%v", e.WrapError(op, "value is invalid", err))
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}
	if err := h.Storage.Replace(r.Context(), name, val); err != nil {
		errStr := fmt.Sprintf("%v", e.WrapError(op, "failed to replace metric", err))
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (h *NotImplementedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.NotImplementedHandler.ServeHTTP()"

	t := chi.URLParam(r, "type")
	if t != "gauge" && t != "counter" {
		errStr := fmt.Sprintf("%v", e.WrapError(op, "unknown metric type", nil))
		http.Error(w, errStr, http.StatusNotImplemented)
		return
	}
}
