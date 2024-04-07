package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	errors2 "github.com/arxon31/metrics-collector/internal/repository/errs"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
)

type GetJSONMetric Handler

func (h *GetJSONMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetJSONMetric.ServeHTTP()"

	w.Header().Set("Content-Type", "application/json")

	var m metric.MetricDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, e.WrapString(op, "failed to decode metric", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	switch m.MType {
	case "gauge":
		val, err := h.Provider.GaugeValue(r.Context(), m.ID)
		if err != nil {
			if errors.Is(err, errors2.ErrMetricNotFound) {
				http.Error(w, e.WrapString(op, "metric does not exist", err), http.StatusNotFound)
			} else {
				http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusInternalServerError)
			}
			return
		}
		m.Value = &val
	case "counter":
		val, err := h.Provider.CounterValue(r.Context(), m.ID)
		if err != nil {
			if errors.Is(err, errors2.ErrMetricNotFound) {
				http.Error(w, e.WrapString(op, "metric does not exist", err), http.StatusNotFound)
			} else {
				http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusInternalServerError)
			}
			return
		}
		m.Delta = &val
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
