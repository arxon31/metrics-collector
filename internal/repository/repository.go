// Package repository provides repository interface
package repository

import (
	"context"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/internal/repository/postgres"
	"github.com/arxon31/metrics-collector/pkg/metric"
)

type Repository interface {
	// Replace replaces gauge metric value
	Replace(ctx context.Context, name string, value float64) error
	// Count increases counter metric value
	Count(ctx context.Context, name string, value int64) error
	// GaugeValue returns gauge metric value
	GaugeValue(ctx context.Context, name string) (float64, error)
	// CounterValue returns counter metric value
	CounterValue(ctx context.Context, name string) (int64, error)
	// Values returns all metrics values
	Values(ctx context.Context) (string, error)
	// Dump dumps all metrics in file
	Dump(ctx context.Context, path string) error
	// Restore restores all metrics from file
	Restore(ctx context.Context, path string) error
	// StoreBatch stores batch of metrics
	StoreBatch(ctx context.Context, metrics []metric.MetricDTO) error
	// Ping pings repository
	Ping() error
}

// New creates new repository.
// If dsn is empty, returns memory storage
func New(dsn string, logger *zap.SugaredLogger) (Repository, error) {
	if dsn == "" {
		return memory.NewMapStorage(), nil
	} else {
		psql, err := postgres.NewPostgres(dsn, logger)
		if err != nil {
			return nil, err
		}
		return psql, nil
	}
}
