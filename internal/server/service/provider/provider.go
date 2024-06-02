package provider

import (
	"context"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/entity"
)

type provider interface {
	Gauge(ctx context.Context, name string) (float64, error)
	Counter(ctx context.Context, name string) (int64, error)
	Metrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type providerService struct {
	provider provider
	logger   *zap.SugaredLogger
}

// NewProviderService initializes a new provider service.
func NewProviderService(provider provider, logger *zap.SugaredLogger) *providerService {
	return &providerService{
		provider: provider,
		logger:   logger,
	}
}

// GetCounterValue returns value of counter by name
func (s *providerService) GetCounterValue(ctx context.Context, name string) (int64, error) {
	val, err := s.provider.Counter(ctx, name)
	if err != nil {
		return -1, err
	}

	return val, nil
}

// GetGaugeValue returns value of gauge by name
func (s *providerService) GetGaugeValue(ctx context.Context, name string) (float64, error) {
	val, err := s.provider.Gauge(ctx, name)
	if err != nil {
		return -1, err
	}
	return val, nil
}

// GetMetrics returns all metrics
func (s *providerService) GetMetrics(ctx context.Context) ([]entity.MetricDTO, error) {
	vals, err := s.provider.Metrics(ctx)
	if err != nil {
		return nil, err
	}
	validMetrics := make([]entity.MetricDTO, 0, len(vals))

	for _, metric := range vals {
		err = metric.Validate()
		if err != nil {
			s.logger.Error(err)
		}
		validMetrics = append(validMetrics, metric)
	}

	return validMetrics, nil
}
