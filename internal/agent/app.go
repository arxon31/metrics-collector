// Package agent collects in itself poller generator and reporter and works with poll and report timeout
package agent

import (
	"github.com/arxon31/metrics-collector/internal/agent/config"
	"github.com/arxon31/metrics-collector/internal/entity"
	"go.uber.org/zap"
	"net/http"
)

type poller interface {
	Poll() *entity.Metrics
}

type RequestGenerator interface {
	Generate(*entity.Metrics) chan *http.Request
}

type MetricsReporter interface {
	Report(requests chan *http.Request)
}

type Agent struct {
	poller          MetricsPoller
	generator       RequestGenerator
	reporter        MetricsReporter
	logger          *zap.SugaredLogger
	config          *config.Config
	metricsToReport *entity.Metrics
}
