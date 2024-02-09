package generator

import (
	"github.com/arxon31/metrics-collector/pkg/metric"
	"net/http"
)

type Generator interface {
	Generate(*metric.Metrics) chan *http.Request
}
