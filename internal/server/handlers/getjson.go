package handlers

import (
	"encoding/json"
	"errors"
	"github.com/arxon31/metrics-collector/internal/entity"
	errors2 "github.com/arxon31/metrics-collector/internal/repository/repoerr"
	"net/http"

	"github.com/arxon31/metrics-collector/pkg/e"
)

// GetJSONMetric implements http.Handler which gets the metric from repository and returns it in JSON
type GetJSONMetric Handler

func (h *GetJSONMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetJSONMetric.ServeHTTP()"

	w.Header().Set("Content-Type", "application/json")

	var m entity.MetricDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, e.WrapString(op, "failed to decode metric", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	switch m.MetricType {
	case "gauge":
		val, err := h.Provider.GaugeValue(r.Context(), m.Name)
		if err != nil {
			if errors.Is(err, errors2.ErrMetricNotFound) {
				http.Error(w, e.WrapString(op, "metric does not exist", err), http.StatusNotFound)
			} else {
				http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusInternalServerError)
			}
			return
		}
		m.Gauge = &val
	case "counter":
		val, err := h.Provider.CounterValue(r.Context(), m.Name)
		if err != nil {
			if errors.Is(err, errors2.ErrMetricNotFound) {
				http.Error(w, e.WrapString(op, "metric does not exist", err), http.StatusNotFound)
			} else {
				http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusInternalServerError)
			}
			return
		}
		m.Counter = &val
	default:
		http.Error(w, e.WrapString(op, "unknown metric type", nil), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, e.WrapString(op, "failed to marshal metric", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
