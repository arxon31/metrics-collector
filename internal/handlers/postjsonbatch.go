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

	err := h.Storage.StoreBatch(r.Context(), ms)
	if err != nil {
		h.Logger.Errorln(e.WrapString(op, "failed to store batch of metrics", err))
		http.Error(w, e.WrapString(op, "failed to store batch of metrics", err), http.StatusInternalServerError)
	}

	resp, err := json.Marshal(ms)
	if err != nil {
		http.Error(w, e.WrapString(op, "failed to marshal metric", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}
