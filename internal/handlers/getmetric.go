package handlers

import (
	"errors"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type GetMetricHandler Handler

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetMetricHandler.ServeHTTP()"

	t := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch t {
	case "gauge":
		value, err := h.Provider.GaugeValue(r.Context(), name)
		if err != nil {
			if errors.Is(err, mem.ErrIsNotFound) {
				http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusNotFound)
				return
			}

		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%v", value)))

	case "counter":
		value, err := h.Provider.CounterValue(r.Context(), name)
		if err != nil {
			if errors.Is(err, mem.ErrIsNotFound) {
				http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusNotFound)
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%v", value)))

	default:
		errStr := fmt.Sprintf("%v", e.WrapError(op, "unknown metric type", nil))
		http.Error(w, errStr, http.StatusNotFound)
	}

}

type GetMetricsHandler Handler

func (h *GetMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetMetricsHandler.ServeHTTP()"
	body, err := h.Provider.Values(r.Context())
	if err != nil {
		errStr := fmt.Sprintf("%v", e.WrapError(op, "failed to get metrics", err))
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}
