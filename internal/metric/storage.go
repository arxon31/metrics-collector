package metric

import "github.com/arxon31/metrics-collector/internal/metric/metrics"

type Storage struct {
	Storage map[string]metrics.Metric
}

func NewMetricStorage() Storage {
	return Storage{
		Storage: make(map[string]metrics.Metric),
	}
}

func (s Storage) Update(name string, newVal interface{}) error {
	s.Storage[name].SetValue(newVal)
	return nil
}

func (s Storage) GetMetric(name string) interface{} {
	metric, ok := s.Storage[name]
	if !ok {
		return nil
	}
	return metric.GetValue()
}

func (s Storage) IsExist(name string) bool {
	_, ok := s.Storage[name]
	return ok
}
