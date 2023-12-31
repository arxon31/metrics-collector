package handlers

import (
	"fmt"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"github.com/go-chi/chi/v5"
	"net/http"
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

// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func (h *PostCounterMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostCounterMetric.ServeHTTP()"

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var counter metric.Counter
	val, err := counter.CounterFromString(value)
	if err != nil {
		errStr := fmt.Sprintf("%v", e.Wrap(op, "value is invalid", err))
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}

	if err := h.Storage.Count(r.Context(), name, val); err != nil {
		errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to replace metric", err))
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

}

// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func (h *PostGaugeMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostGaugeMetric.ServeHTTP()"

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	var gauge metric.Gauge

	val, err := gauge.GaugeFtomString(value)
	if err != nil {
		errStr := fmt.Sprintf("%v", e.Wrap(op, "value is invalid", err))
		http.Error(w, errStr, http.StatusBadRequest)
		return
	}
	if err := h.Storage.Replace(r.Context(), name, val); err != nil {
		errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to replace metric", err))
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
		errStr := fmt.Sprintf("%v", e.Wrap(op, "unknown metric type", nil))
		http.Error(w, errStr, http.StatusNotImplemented)
		return
	}
}
