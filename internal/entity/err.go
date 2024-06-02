package entity

import "errors"

var (
	ErrMetricName     = errors.New("metric name is empty")
	ErrMetricType     = errors.New("metric type is empty")
	ErrCounterValue   = errors.New("counter value is empty")
	ErrGaugeValue     = errors.New("gauge value is empty")
	ErrNoValue        = errors.New("no value provided")
	ErrMultipleValues = errors.New("multiple values provided")
)
