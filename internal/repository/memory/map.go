package memory

import (
	"context"
	"sync"

	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/arxon31/metrics-collector/internal/repository/repoerr"
)

type MapStorage struct {
	rw     *sync.RWMutex
	gauges map[string]float64
	counts map[string]int64
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		rw:     &sync.RWMutex{},
		gauges: make(map[string]float64),
		counts: make(map[string]int64),
	}
}

func (s *MapStorage) StoreGauge(_ context.Context, name string, value float64) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.gauges[name] = value
	return nil
}

func (s *MapStorage) StoreCounter(_ context.Context, name string, value int64) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	if _, ok := s.counts[name]; !ok {
		s.counts[name] = value
		return nil
	}
	s.counts[name] += value
	return nil
}

func (s *MapStorage) Gauge(_ context.Context, name string) (float64, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if val, ok := s.gauges[name]; ok {
		return val, nil
	}
	return -1, repoerr.ErrMetricNotFound
}

func (s *MapStorage) Counter(_ context.Context, name string) (int64, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if val, ok := s.counts[name]; ok {
		return val, nil
	}
	return -1, repoerr.ErrMetricNotFound
}

func (s *MapStorage) Metrics(_ context.Context) ([]entity.MetricDTO, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	metrics := make([]entity.MetricDTO, 0, len(s.gauges)+len(s.counts))

	for name, value := range s.gauges {
		val := value
		metrics = append(metrics, entity.MetricDTO{
			Name:       name,
			MetricType: entity.GaugeType,
			Gauge:      &val,
		})
	}

	for name, value := range s.counts {
		val := value
		metrics = append(metrics, entity.MetricDTO{
			Name:       name,
			MetricType: entity.CounterType,
			Counter:    &val,
		})
	}

	return metrics, nil
}

func (s *MapStorage) StoreBatch(_ context.Context, metrics []entity.MetricDTO) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	for _, m := range metrics {
		switch m.MetricType {
		case entity.GaugeType:
			s.gauges[m.Name] = *m.Gauge

		case entity.CounterType:
			if _, ok := s.counts[m.Name]; !ok {
				s.counts[m.Name] = *m.Counter
				return nil
			}
			s.counts[m.Name] += *m.Counter
		}
	}
	return nil

}

func (s *MapStorage) Ping() error {
	return nil
}
