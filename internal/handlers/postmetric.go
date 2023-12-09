package handlers

import (
	"context"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/storage"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type PostMetrics struct {
	Storage storage.Storage
}

func (h *PostMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostMetrics.ServeHTTP()"

	//--httpserver.Params--|httpserver.GaugePath|------------params-------------|
	//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>

	t := chi.URLParam(r, "type")
	if t != "gauge" && t != "counter" {
		errStr := fmt.Sprintf("%v", e.Wrap(op, "unknown metric type", nil))
		http.Error(w, errStr, http.StatusNotImplemented)
		return
	}

	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")

	switch t {
	case "gauge":
		var gauge metric.Gauge

		val, err := gauge.Validate(value)
		if err != nil {
			errStr := fmt.Sprintf("%v", e.Wrap(op, "value is invalid", err))
			http.Error(w, errStr, http.StatusBadRequest)
			return
		}
		if err := h.Storage.Replace(context.Background(), name, val.(float64)); err != nil {
			errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to replace metric", err))
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	case "counter":
		var counter metric.Counter
		val, err := counter.Validate(value)
		if err != nil {
			errStr := fmt.Sprintf("%v", e.Wrap(op, "value is invalid", err))
			http.Error(w, errStr, http.StatusBadRequest)
			return
		}

		if err := h.Storage.Count(context.Background(), name, val.(int64)); err != nil {
			errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to replace metric", err))
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}

}
