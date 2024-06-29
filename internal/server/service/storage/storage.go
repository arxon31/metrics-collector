package storage

import (
	"context"
	"github.com/arxon31/metrics-collector/pkg/logger"

	"github.com/arxon31/metrics-collector/internal/entity"
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
	repo storage
}

// NewStorageService initializes a new storage service.
func NewStorageService(repo storage) *storageService {
	return &storageService{
		repo: repo,
	}
}

// SaveGaugeMetric saves the metric in repo
func (s *storageService) SaveGaugeMetric(ctx context.Context, metric entity.MetricDTO) error {
	err := metric.Validate()
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	err = s.repo.StoreGauge(ctx, metric.Name, *metric.Gauge)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	return nil
}

// SaveCounterMetric saves the metric in repo
func (s *storageService) SaveCounterMetric(ctx context.Context, metric entity.MetricDTO) error {
	err := metric.Validate()
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	err = s.repo.StoreCounter(ctx, metric.Name, *metric.Counter)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	return nil
}

// SaveBatchMetrics saves metrics in repo
func (s *storageService) SaveBatchMetrics(ctx context.Context, metrics []entity.MetricDTO) error {
	validMetrics := make([]entity.MetricDTO, 0, len(metrics))

	for _, metric := range metrics {
		err := metric.Validate()
		if err != nil {
			logger.Logger.Error(err)
		}
		validMetrics = append(validMetrics, metric)
	}

	err := s.repo.StoreBatch(ctx, validMetrics)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}
