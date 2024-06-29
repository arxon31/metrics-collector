package entity

import "fmt"

//easyjson:json
type MetricDTOs []MetricDTO

type MetricDTO struct {
	Name       string   `json:"id"`
	MetricType string   `json:"type"`
	Counter    *int64   `json:"delta,omitempty"`
	Gauge      *float64 `json:"value,omitempty"`
}

func (m *MetricDTO) Validate() error {
	if m.Name == "" {
		return ErrMetricName
	}
	if m.MetricType == "" || (m.MetricType != GaugeType && m.MetricType != CounterType) {
		return fmt.Errorf("%s:%w", m.Name, ErrMetricType)
	}

	if m.MetricType == CounterType && m.Counter == nil {
		return fmt.Errorf("%s:%w", m.Name, ErrCounterValue)
	}
	if m.MetricType == GaugeType && m.Gauge == nil {
		return fmt.Errorf("%s:%w", m.Name, ErrGaugeValue)
	}

	if m.Gauge == nil && m.Counter == nil {
		return fmt.Errorf("%s:%w", m.Name, ErrNoValue)
	}

	if m.Gauge != nil && m.Counter != nil {
		return fmt.Errorf("%s:%w", m.Name, ErrMultipleValues)
	}

	return nil
}
