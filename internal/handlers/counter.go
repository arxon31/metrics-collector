package handlers

import (
	"context"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/storage"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"net/http"
	"strings"
)

type CounterHandler struct {
	Storage storage.Storage
}

func (h *CounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.CounterHandler.ServeHTTP()"
	//--httpserver.Params--|httpserver.GaugePath|------------params-------------|
	//http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>

	// Отрезаем дефолтный путь мультиплексора
	raw := r.URL.Path[len("/update/counter/"):]
	params := strings.Split(raw, "/")
	if len(params) != 2 {
		http.Error(w, "Not enough params for request", http.StatusNotFound)
		return
	}
	name := params[0]

	var counter metric.Counter
	val, err := counter.Validate(params[1])
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
