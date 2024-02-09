package poller

import (
	"github.com/arxon31/metrics-collector/pkg/metric"
)

type Poller interface {
	Poll() *metric.Metrics
}
