// Package repository provides repository interface
package repository

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/entity"
)

type Repository interface {
	// StoreGauge replaces gauge metric value
	StoreGauge(ctx context.Context, name string, value float64) error
	// StoreCounter increases counter metric value
	StoreCounter(ctx context.Context, name string, value int64) error
	// Gauge returns gauge metric value
	Gauge(ctx context.Context, name string) (float64, error)
	// Counter returns counter metric value
	Counter(ctx context.Context, name string) (int64, error)
	// Metrics returns all metrics values
	Metrics(ctx context.Context) (string, error)
	// StoreBatch stores batch of metrics
	StoreBatch(ctx context.Context, metrics []entity.MetricDTO) error
}
