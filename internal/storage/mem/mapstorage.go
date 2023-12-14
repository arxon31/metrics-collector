package mem

import (
	"context"
	"errors"
	"fmt"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"strings"
	"sync"
)

var ErrIsNotFound = errors.New("not found")

type MapStorage struct {
	rw *sync.RWMutex
	ms *metric.Metrics
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		rw: &sync.RWMutex{},
		ms: metric.New(),
	}
}

func (s *MapStorage) Replace(ctx context.Context, name string, value float64) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.ms.Gauges[metric.Name(name)] = metric.Gauge(value)
	return nil
}

func (s *MapStorage) Count(ctx context.Context, name string, value int64) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	if _, ok := s.ms.Counters[metric.Name(name)]; !ok {
		s.ms.Counters[metric.Name(name)] = metric.Counter(value)
		return nil
	}
	s.ms.Counters[metric.Name(name)] += metric.Counter(value)
	return nil
}

func (s *MapStorage) GaugeValue(ctx context.Context, name string) (float64, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if val, ok := s.ms.Gauges[metric.Name(name)]; ok {
		return float64(val), nil
	}
	return 0, ErrIsNotFound
}

func (s *MapStorage) CounterValue(ctx context.Context, name string) (int64, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if val, ok := s.ms.Counters[metric.Name(name)]; ok {
		return int64(val), nil
	}
	return 0, ErrIsNotFound
}

func (s *MapStorage) Values(ctx context.Context) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	var body strings.Builder

	for name, value := range s.ms.Gauges {
		body.WriteString(fmt.Sprintf("%v %v\n", name, value))
	}

	for name, value := range s.ms.Counters {
		body.WriteString(fmt.Sprintf("%v %v\n", name, value))
	}

	return body.String(), nil

}
