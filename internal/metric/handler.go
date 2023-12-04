package metric

import (
	"errors"
	"github.com/arxon31/metrics-collector/internal/handlers"
	"github.com/arxon31/metrics-collector/internal/metric/metrics"
	"github.com/arxon31/metrics-collector/internal/service"
	"net/http"
	"strconv"
	"strings"
)

const (
	updateURL         = "/update/"
	GaugeMetricType   = "gauge"
	CounterMetricType = "counter"
	errBadRequest     = "not enough params"
	errInvalidType    = "invalid metric type"
	errInvalidValue   = "invalid metric value"
)

type handler struct {
	service service.Service
}

func NewHandler(service service.Service) handlers.Handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(updateURL, h.Update)
}

func (h *handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metricPath := r.URL.Path
	raw := metricPath[len(updateURL):]

	params := strings.Split(raw, "/")
	if len(params) != 3 {
		http.Error(w, errBadRequest, http.StatusNotFound)
		return
	}

	metric, err := createMetric(params)
	if err != nil {
		switch err.Error() {
		case errInvalidValue:
			http.Error(w, errInvalidValue, http.StatusBadRequest)
			return
		case errInvalidType:
			http.Error(w, errInvalidType, http.StatusBadRequest)
			return
		}

		if err := h.service.UpdateMetric(metric); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func createMetric(params []string) (metrics.Metric, error) {
	if params[0] != GaugeMetricType && params[0] != CounterMetricType {
		return nil, errors.New(errInvalidType)
	}
	t := params[0]
	name := params[1]
	if t == GaugeMetricType {
		value, err := strconv.ParseFloat(params[2], 64)
		if err != nil {
			return nil, errors.New(errInvalidValue)
		}
		return metrics.NewGaugeMetric(t, name, value), nil
	} else {
		value, err := strconv.ParseInt(params[2], 10, 64)
		if err != nil {
			return nil, errors.New(errInvalidValue)
		}
		return metrics.NewCounterMetric(t, name, value), nil
	}

}
