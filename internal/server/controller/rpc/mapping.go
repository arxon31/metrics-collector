package rpc

import (
	"github.com/arxon31/metrics-proto/pkg/protobuf/metrics"

	"github.com/arxon31/metrics-collector/internal/entity"
)

func mapDTOToProto(m *entity.MetricDTO) *metrics.Metric {
	protoMetric := &metrics.Metric{
		Name: m.Name,
	}

	switch m.MetricType {
	case entity.GaugeType:
		protoMetric.MetricType = metrics.MetricType_GAUGE
		protoMetric.Value = &metrics.Metric_Gauge{Gauge: *m.Gauge}
	case entity.CounterType:
		protoMetric.MetricType = metrics.MetricType_COUNTER
		protoMetric.Value = &metrics.Metric_Counter{Counter: *m.Counter}
	}

	return protoMetric
}

func mapProtoToDTO(mProto *metrics.Metric) *entity.MetricDTO {
	metricDTO := &entity.MetricDTO{
		Name: mProto.GetName(),
	}

	switch mProto.MetricType {
	case metrics.MetricType_GAUGE:
		metricDTO.MetricType = entity.GaugeType
		gauge := mProto.GetGauge()
		metricDTO.Gauge = &gauge
	case metrics.MetricType_COUNTER:
		metricDTO.MetricType = entity.CounterType
		counter := mProto.GetCounter()
		metricDTO.Counter = &counter
	}

	return metricDTO
}
