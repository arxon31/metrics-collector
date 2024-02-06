package mem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"io"
	"os"
	"path/filepath"
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

func (s *MapStorage) Dump(ctx context.Context, path string) error {
	// create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := s.toJSON(ctx)
	if err != nil {
		return err
	}
	n, err := file.Write([]byte(data))
	if err != nil {
		return err

	}
	if n < len(data) {
		return err
	}
	return nil
}

func (s *MapStorage) toJSON(ctx context.Context) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var res strings.Builder
	last := 0
	res.WriteString("[")
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
		res.WriteString(",\n")
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
		last++
		if last != len(s.ms.Counters) {
			res.WriteString(",\n")
		}

	}
	res.WriteString("\n]")

	return res.String(), nil
}

func (s *MapStorage) Restore(ctx context.Context, path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return ErrIsNotFound
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err

	}
	err = s.fromJSON(ctx, string(data))
	if err != nil {
		return err
	}
	return nil
}

func (s *MapStorage) fromJSON(ctx context.Context, values string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	metrics := make([]metric.MetricDTO, metric.CounterCount+metric.GaugeCount)
	err := json.Unmarshal([]byte(values), &metrics)
	if err != nil {
		return err
	}
	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			s.ms.Gauges[metric.Name(m.ID)] = metric.Gauge(*m.Value)
		case "counter":
			s.ms.Counters[metric.Name(m.ID)] = metric.Counter(*m.Delta)
		}
	}
	return nil
}

func (s *MapStorage) Ping() error {
	return nil
}

func (s *MapStorage) StoreBatch(ctx context.Context, metrics []metric.MetricDTO) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	for _, m := range metrics {
		if m.MType == "gauge" {
			metricVal := *m.Value
			s.ms.Gauges[metric.Name(m.ID)] = metric.Gauge(metricVal)
		} else {
			if _, ok := s.ms.Counters[metric.Name(m.ID)]; !ok {
				s.ms.Counters[metric.Name(m.ID)] = metric.Counter(*m.Delta)
				return nil
			}
			s.ms.Counters[metric.Name(m.ID)] += metric.Counter(*m.Delta)
		}
	}
	return nil
}
