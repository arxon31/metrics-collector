package handlers

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/entity"

	"go.uber.org/zap"
)

type MetricProvider interface {
	GaugeValue(ctx context.Context, name string) (float64, error)
	CounterValue(ctx context.Context, name string) (int64, error)
	Values(ctx context.Context) (string, error)
}

type MetricCollector interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
	StoreBatch(ctx context.Context, metrics []entity.MetricDTO) error
}

type Pinger interface {
	Ping() error
}

type Handler struct {
	Storage  MetricCollector
	Provider MetricProvider
	Logger   *zap.SugaredLogger
}

type CustomHandler struct {
	Pinger
}
