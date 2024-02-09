package repository

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/repository/memory"
	"github.com/arxon31/metrics-collector/internal/repository/postgres"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"go.uber.org/zap"
)

type Repository interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
	GaugeValue(ctx context.Context, name string) (float64, error)
	CounterValue(ctx context.Context, name string) (int64, error)
	Values(ctx context.Context) (string, error)
	Dump(ctx context.Context, path string) error
	Restore(ctx context.Context, path string) error
	StoreBatch(ctx context.Context, metrics []metric.MetricDTO) error
	Ping() error
}

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
