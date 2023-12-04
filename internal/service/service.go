package service

import (
	"errors"
	"github.com/arxon31/metrics-collector/internal/metric/metrics"
	"github.com/arxon31/metrics-collector/internal/storage"
)

const (
	unsupportedType   = "unsupported metric type"
	gaugeMetricType   = "gauge"
	counterMetricType = "counter"
)

type Service struct {
	storage storage.Storage
}

func NewService(storage storage.Storage) Service {
	return Service{
		storage: storage,
	}
}

func (s *Service) UpdateMetric(metric metrics.Metric) error {
	if metric.GetType() == gaugeMetricType {
		err := s.updateGaugeMetric(metric)
		if err != nil {
			return err
		}
	} else if metric.GetType() == counterMetricType {
		err := s.updateCounterMetric(metric)
		if err != nil {
			return err
		}
	}

	return errors.New(unsupportedType)

}

func (s *Service) updateGaugeMetric(metric metrics.Metric) error {
	name := metric.GetName()
	newVal := metric.GetValue()
	if err := s.storage.Update(name, newVal); err != nil {
		return err
	}
	return nil

}

func (s *Service) updateCounterMetric(metric metrics.Metric) error {
	var oldVal interface{}
	name := metric.GetName()
	if s.storage.IsExist(name) {
		oldVal = s.storage.GetMetric(name)
	}
	val := metric.GetValue().(int64) + oldVal.(int64)

	if err := s.storage.Update(name, val); err != nil {
		return err
	}

	return nil

}
