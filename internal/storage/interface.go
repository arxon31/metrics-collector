package storage

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/arxon31/metrics-collector/internal/storage/postgres"
	"go.uber.org/zap"
)

type Storage interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
	GaugeValue(ctx context.Context, name string) (float64, error)
	CounterValue(ctx context.Context, name string) (int64, error)
	Values(ctx context.Context) (string, error)
	Dump(ctx context.Context, path string) error
	Restore(ctx context.Context, path string) error
}

func New(dsn string, logger *zap.SugaredLogger) (Storage, error) {
	if dsn == "" {
		return mem.NewMapStorage(), nil
	} else {
		psql, err := postgres.NewPostgres(dsn, logger)
		if err != nil {
			return nil, err
		}
		return psql, nil
	}
}
