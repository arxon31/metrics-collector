package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arxon31/metrics-proto/pkg/protobuf/metrics"

	"github.com/arxon31/metrics-collector/internal/entity"
)

type storageService interface {
	SaveGaugeMetric(ctx context.Context, metric entity.MetricDTO) error
	SaveCounterMetric(ctx context.Context, metric entity.MetricDTO) error
	SaveBatchMetrics(ctx context.Context, metrics []entity.MetricDTO) error
}

type providerService interface {
	GetGaugeValue(ctx context.Context, name string) (float64, error)
	GetCounterValue(ctx context.Context, name string) (int64, error)
	GetMetrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type server struct {
	storage  storageService
	provider providerService
	metrics.UnimplementedMetricsCollectorServer
}

func NewGRPCController(storage storageService, provider providerService) *server {
	return &server{storage: storage, provider: provider}
}

func (s *server) GetMetric(ctx context.Context, mGet *metrics.GetMetricRequest) (*metrics.GetMetricResponse, error) {
	switch mGet.GetType() {
	case metrics.MetricType_COUNTER:
		val, err := s.provider.GetCounterValue(ctx, mGet.GetName())
		if err != nil {
			return nil, err
		}
		return &metrics.GetMetricResponse{
			Metric: &metrics.Metric{
				Name:       mGet.GetName(),
				MetricType: metrics.MetricType_COUNTER,
				Value:      &metrics.Metric_Counter{Counter: val},
			},
		}, nil

	case metrics.MetricType_GAUGE:
		val, err := s.provider.GetGaugeValue(ctx, mGet.GetName())
		if err != nil {
			return nil, err
		}
		return &metrics.GetMetricResponse{
			Metric: &metrics.Metric{
				Name:       mGet.GetName(),
				MetricType: metrics.MetricType_GAUGE,
				Value:      &metrics.Metric_Gauge{Gauge: val},
			},
		}, nil

	}

	return nil, status.Error(codes.InvalidArgument, "unsupported metric type")
}
func (s *server) GetMetrics(ctx context.Context, msGet *metrics.GetMetricsRequest) (*metrics.GetMetricsResponse, error) {
	ms, err := s.provider.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}

	req := &metrics.GetMetricsResponse{Metrics: make([]*metrics.Metric, len(ms))}

	for i, metric := range ms {
		req.Metrics[i] = mapDTOToProto(&metric)
	}

	return req, nil

}
func (s *server) AddMetric(ctx context.Context, mAdd *metrics.AddMetricRequest) (*metrics.AddMetricResponse, error) {
	mDTO := mapProtoToDTO(mAdd.Metric)

	switch mDTO.MetricType {
	case entity.GaugeType:
		err := s.storage.SaveGaugeMetric(ctx, *mDTO)
		if err != nil {
			return nil, err
		}

		val, err := s.provider.GetGaugeValue(ctx, mDTO.Name)
		if err != nil {
			return nil, err
		}

		return &metrics.AddMetricResponse{
			Metric: &metrics.Metric{
				Name:       mDTO.Name,
				MetricType: metrics.MetricType_GAUGE,
				Value:      &metrics.Metric_Gauge{Gauge: val},
			}}, nil

	case entity.CounterType:
		err := s.storage.SaveCounterMetric(ctx, *mDTO)
		if err != nil {
			return nil, err
		}

		val, err := s.provider.GetCounterValue(ctx, mDTO.Name)
		if err != nil {
			return nil, err
		}

		return &metrics.AddMetricResponse{
			Metric: &metrics.Metric{
				Name:       mDTO.Name,
				MetricType: metrics.MetricType_COUNTER,
				Value:      &metrics.Metric_Counter{Counter: val},
			}}, nil
	}

	return nil, status.Error(codes.InvalidArgument, "unsupported metric type")

}
func (s *server) AddMetrics(ctx context.Context, msAdd *metrics.AddMetricsRequest) (*metrics.AddMetricsResponse, error) {
	msDTO := make([]entity.MetricDTO, len(msAdd.Metrics))
	msProto := make([]*metrics.Metric, 0, len(msAdd.Metrics))

	for i, mProto := range msAdd.Metrics {
		mDTO := mapProtoToDTO(mProto)
		msDTO[i] = *mDTO
	}

	err := s.storage.SaveBatchMetrics(ctx, msDTO)
	if err != nil {
		return nil, err
	}

	ms, err := s.provider.GetMetrics(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range ms {
		msProto = append(msProto, mapDTOToProto(&m))
	}

	return &metrics.AddMetricsResponse{Metrics: msProto}, nil
}
