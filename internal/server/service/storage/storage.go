package storage

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/entity"
	"go.uber.org/zap"
)

type storage interface {
	// StoreGauge replaces gauge metric value
	StoreGauge(ctx context.Context, name string, value float64) error
	// StoreCounter increases counter metric value
	StoreCounter(ctx context.Context, name string, value int64) error
	// StoreBatch stores batch of metrics
	StoreBatch(ctx context.Context, metrics []entity.MetricDTO) error
}

type storageService struct {
	repo   storage
	logger *zap.SugaredLogger
}

func NewStorageService(repo storage, logger *zap.SugaredLogger) *storageService {
	return &storageService{
		repo:   repo,
		logger: logger,
	}
}

func (s *storageService) SaveGaugeMetric(ctx context.Context, metric entity.MetricDTO) error {
	err := metric.Validate()
	if err != nil {
		s.logger.Error(err)
		return err
	}

	err = s.repo.StoreGauge(ctx, metric.Name, *metric.Gauge)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	return nil
}

func (s *storageService) SaveCounterMetric(ctx context.Context, metric entity.MetricDTO) error {
	err := metric.Validate()
	if err != nil {
		s.logger.Error(err)
		return err
	}

	err = s.repo.StoreCounter(ctx, metric.Name, *metric.Counter)
	if err != nil {
		s.logger.Error(err)
		return err
	}

	return nil
}

func (s *storageService) SaveBatchMetrics(ctx context.Context, metrics []entity.MetricDTO) error {
	validMetrics := make([]entity.MetricDTO, 0, len(metrics))

	for _, metric := range metrics {
		err := metric.Validate()
		if err != nil {
			s.logger.Error(err)
		}
		validMetrics = append(validMetrics, metric)
	}

	err := s.repo.StoreBatch(ctx, validMetrics)
	if err != nil {
		s.logger.Error(err)
		return err
	}
	return nil
}
