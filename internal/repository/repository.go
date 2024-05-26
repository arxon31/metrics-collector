// Package repository provides repository interface
package repository

import (
	"context"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/internal/repository/postgres"
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
	Metrics(ctx context.Context) ([]entity.MetricDTO, error)
	// StoreBatch stores batch of metrics
	StoreBatch(ctx context.Context, metrics []entity.MetricDTO) error
	// Ping checks connection
	Ping() error
}

func New(url string, logger *zap.SugaredLogger) (Repository, error) {
	if url == "" {
		return memory.NewMapStorage(), nil
	} else {
		return postgres.NewPostgres(url, logger)
	}
}
