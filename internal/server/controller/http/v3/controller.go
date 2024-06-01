package v3

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/arxon31/metrics-collector/internal/entity"
)

const (
	saveJSONMetricsURL = "/updates/"
	getJSONMetricsURL  = "/"
	pingDBURL          = "/ping"
)

//go:generate moq -out storageService_moq_test.go . storageService
type storageService interface {
	SaveBatchMetrics(ctx context.Context, metrics []entity.MetricDTO) error
}

//go:generate moq -out providerService_moq_test.go . providerService
type providerService interface {
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetMetrics(ctx context.Context) ([]entity.MetricDTO, error)
}

//go:generate moq -out pingerService_moq_test.go . pingerService
type pingerService interface {
	PingDB() error
}

type v3 struct {
	store    storageService
	provider providerService
	pinger   pingerService
}

func NewController(store storageService, provider providerService, pinger pingerService) *v3 {
	return &v3{
		store:    store,
		provider: provider,
		pinger:   pinger,
	}
}

func (v *v3) Register(h *chi.Mux) {
	h.Get(pingDBURL, v.pingDB)
	h.Get(getJSONMetricsURL, v.getJSONMetrics)
	h.Post(saveJSONMetricsURL, v.saveJSONMetrics)
}

func (v *v3) pingDB(w http.ResponseWriter, r *http.Request) {
	err := v.pinger.PingDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (v *v3) saveJSONMetrics(w http.ResponseWriter, r *http.Request) {
	ms := make([]entity.MetricDTO, 0)

	if err := json.NewDecoder(r.Body).Decode(&ms); err != nil {
		http.Error(w, fmt.Sprintf("can not decode metrics: %s", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := v.store.SaveBatchMetrics(r.Context(), ms)
	if err != nil {
		http.Error(w, "failed to store batch of metrics", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (v *v3) getJSONMetrics(w http.ResponseWriter, r *http.Request) {
	ms, err := v.provider.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("can not get metrics: %s", err), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(ms)
	if err != nil {
		http.Error(w, fmt.Sprintf("can not marshal metrics: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
