package handlers

import (
	"encoding/json"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"net/http"
)

type GetJSONMetric Handler

func (h *GetJSONMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetJSONMetric.ServeHTTP()"

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
			http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusInternalServerError)
			return
		}
		m.Value = &val
	case "counter":
		val, err := h.Provider.CounterValue(r.Context(), m.ID)
		if err != nil {
			http.Error(w, e.WrapString(op, "failed to get metric", err), http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
