package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/entity"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"

	"github.com/arxon31/metrics-collector/pkg/e"
)

type PostJSONMetric Handler

// PostJSONMetric implements http.Handler which gets the metric JSON from body and stores it in repository
func (h *PostJSONMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.PostJSONMetric.ServeHTTP()"

	w.Header().Set("Content-Type", "application/json")

	var m entity.MetricDTO

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		h.Logger.Errorln(e.WrapString(op, "failed to decode metric", err))
		http.Error(w, e.WrapString(op, "failed to decode metric", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	switch m.MetricType {
	case "gauge":
		if err := h.Storage.Replace(r.Context(), m.Name, *m.Gauge); err != nil {
			if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
				err = retry(3, func() error {
					return h.Storage.Replace(r.Context(), m.Name, *m.Gauge)
				})
				if err != nil {
					http.Error(w, e.WrapString(op, "failed to replace metric with 3 tries", err), http.StatusInternalServerError)
					return
				}
			}
			http.Error(w, e.WrapString(op, "failed to replace metric", err), http.StatusInternalServerError)
			return
		}
	case "counter":
		if err := h.Storage.Count(r.Context(), m.Name, *m.Counter); err != nil {
			if strings.Contains(err.Error(), pgerrcode.UniqueViolation) {
				err = retry(3, func() error {
					return h.Storage.Count(r.Context(), m.Name, *m.Counter)
				})
				if err != nil {
					http.Error(w, e.WrapString(op, "failed to replace metric with 3 tries", err), http.StatusInternalServerError)
					return
				}
			}
			http.Error(w, e.WrapString(op, "failed to count metric", err), http.StatusInternalServerError)
			return
		}
	default:
		h.Logger.Errorln(e.WrapString(op, "unknown metric type", nil))
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

func retry(attempts int, f func() error) (err error) {
	sleep := 1 * time.Second
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep += 2 * time.Second
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
