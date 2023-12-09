package mem

import (
	"context"
	"encoding/json"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"sync"
)

type MapStorage struct {
	rw *sync.Mutex
	ms *metric.Metrics
}

func NewMapStorage() *MapStorage {
	return &MapStorage{
		rw: &sync.Mutex{},
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

func (s *MapStorage) MetricsJSON() (string, error) {
	const op = "MapStorage.MetricsJSON()"
	s.rw.Lock()
	defer s.rw.Unlock()
	j, err := json.Marshal(s.ms)
	if err != nil {
		return "", e.Wrap(op, " can't marshal metrics", err)
	}
	return string(j), nil

}
