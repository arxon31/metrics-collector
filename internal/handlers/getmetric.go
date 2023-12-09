package handlers

import (
	"context"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/storage"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type GetMetricHandler struct {
	Storage storage.Storage
}

func (h *GetMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetMetricHandler.ServeHTTP()"

	t := chi.URLParam(r, "type")
	name := chi.URLParam(r, "name")

	switch t {
	case "gauge":
		value, err := h.Storage.GaugeValue(context.Background(), name)
		if err != nil {
			errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to get metric", err))
			http.Error(w, errStr, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%v", value)))

	case "counter":
		value, err := h.Storage.CounterValue(context.Background(), name)
		if err != nil {
			errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to get metric", err))
			http.Error(w, errStr, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%v", value)))

	default:
		errStr := fmt.Sprintf("%v", e.Wrap(op, "unknown metric type", nil))
		http.Error(w, errStr, http.StatusNotFound)
	}

}
