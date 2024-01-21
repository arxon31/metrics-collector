package handlers

import (
	"encoding/json"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"net/http"
)

type PostJSONBatch Handler

func (h *PostJSONBatch) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostJSONBatch.ServeHTTP()"

	w.Header().Set("Content-Type", "application/json")

	var ms []metric.MetricDTO

	if err := json.NewDecoder(r.Body).Decode(&ms); err != nil {
		h.Logger.Errorln(e.WrapString(op, "failed to decode batch of metrics", err))
		http.Error(w, e.WrapString(op, "failed to decode batch of metrics", err), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	for _, m := range ms {
		switch m.MType {
		case "gauge":
			if err := h.Storage.Replace(r.Context(), m.ID, *m.Value); err != nil {
				h.Logger.Errorln("failed to replace metric from batch", err)
				http.Error(w, e.WrapString(op, "failed to replace metric", err), http.StatusInternalServerError)
				return
			}
		case "counter":
			if err := h.Storage.Count(r.Context(), m.ID, *m.Delta); err != nil {
				h.Logger.Errorln("failed to count metric from batch", err)
				http.Error(w, e.WrapString(op, "failed to count metric", err), http.StatusInternalServerError)
				return
			}
		default:
			h.Logger.Errorln(e.WrapString(op, "unknown metric type", nil))
			http.Error(w, e.WrapString(op, "unknown metric type", nil), http.StatusBadRequest)
			return
		}
	}

	resp, err := json.Marshal(ms)
	if err != nil {
		http.Error(w, e.WrapString(op, "failed to marshal metric", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}
