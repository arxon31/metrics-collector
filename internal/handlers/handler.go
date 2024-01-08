package handlers

import (
	"context"
)

type MetricProvider interface {
	GaugeValue(ctx context.Context, name string) (float64, error)
	CounterValue(ctx context.Context, name string) (int64, error)
	Values(ctx context.Context) (string, error)
}

type MetricCollector interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
}

type Handler struct {
	Storage  MetricCollector
	Provider MetricProvider
}

func NewHandler(storage MetricCollector, provider MetricProvider) *Handler {
	return &Handler{
		Storage:  storage,
		Provider: provider,
	}
}
