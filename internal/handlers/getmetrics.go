package handlers

import (
	"context"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/storage"
	"github.com/arxon31/metrics-collector/pkg/e"
	"net/http"
)

type GetMetricsHandler struct {
	Storage storage.Storage
}

func (h *GetMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.GetMetricsHandler.ServeHTTP()"
	body, err := h.Storage.Values(context.Background())
	if err != nil {
		errStr := fmt.Sprintf("%v", e.Wrap(op, "failed to get metrics", err))
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}
