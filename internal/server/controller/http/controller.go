package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/arxon31/metrics-collector/internal/server/controller/http/middlewares"
	v1 "github.com/arxon31/metrics-collector/internal/server/controller/http/v1"
	v2 "github.com/arxon31/metrics-collector/internal/server/controller/http/v2"
	v3 "github.com/arxon31/metrics-collector/internal/server/controller/http/v3"
)

type storageService interface {
	SaveGaugeMetric(ctx context.Context, metric entity.MetricDTO) error
	SaveCounterMetric(ctx context.Context, metric entity.MetricDTO) error
	SaveBatchMetrics(ctx context.Context, metrics []entity.MetricDTO) error
}

type providerService interface {
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetMetrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type pingerService interface {
	PingDB() error
}

func NewController(handler *chi.Mux, storage storageService, provider providerService, pinger pingerService, logger *zap.SugaredLogger, hashKey string) http.Handler {
	logging := middlewares.NewLoggingMiddleware(logger)
	hashing := middlewares.NewHashingMiddleware(hashKey)
	compressing := middlewares.NewCompressingMiddleware()

	handler.Use(logging.WithLog, hashing.WithHash, compressing.WithCompress)

	sprint1 := v1.NewController(storage, provider)
	sprint1.Register(handler)

	sprint2 := v2.NewController(storage, provider)
	sprint2.Register(handler)

	sprint3 := v3.NewController(storage, provider, pinger)
	sprint3.Register(handler)

	return handler
}
