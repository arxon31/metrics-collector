package mem

import (
	"context"
	"encoding/json"
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

func (s *MapStorage) ValuesJSON(ctx context.Context) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var res strings.Builder
	for name, value := range s.ms.Gauges {
		val := float64(value)
		m := metric.MetricDTO{
			ID:    string(name),
			MType: "gauge",
			Value: &val,
		}
		metricJSON, err := json.Marshal(m)
		if err != nil {
			return "", err
		}
		res.WriteString(string(metricJSON))
		res.WriteString("\n\r")
	}
	for name, value := range s.ms.Counters {
		val := int64(value)
		m := metric.MetricDTO{
			ID:    string(name),
			MType: "counter",
			Delta: &val,
		}
		metricJSON, err := json.Marshal(m)
		if err != nil {
			return "", err
		}
		res.WriteString(string(metricJSON))
		res.WriteString("\n\r")
	}
	return res.String(), nil
}

func (s *MapStorage) RestoreFromJSON(ctx context.Context, values string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	metrics := strings.Split(values, "\n\r")
	for idx, val := range metrics {
		if idx == len(metrics)-1 {
			continue
		}
		m := metric.MetricDTO{}
		err := json.Unmarshal([]byte(val), &m)
		if err != nil {
			return err
		}
		switch m.MType {
		case "gauge":
			s.ms.Gauges[metric.Name(m.ID)] = metric.Gauge(*m.Value)
		case "counter":
			s.ms.Counters[metric.Name(m.ID)] = metric.Counter(*m.Delta)
		}
	}
	return nil
}
