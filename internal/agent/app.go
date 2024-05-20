// Package agent collects in itself poller generator and reporter and works with poll and report timeout
package agent

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/agent/config"
	"github.com/arxon31/metrics-collector/internal/agent/service/generator"
	"github.com/arxon31/metrics-collector/internal/agent/service/poller"
	"github.com/arxon31/metrics-collector/internal/agent/service/reporter"
	"github.com/arxon31/metrics-collector/internal/entity"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type MetricsPoller interface {
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

func New(config *config.Config, logger *zap.SugaredLogger) *Agent {
	return &Agent{
		poller:    poller.New(logger),
		generator: generator.New(config, logger),
		reporter:  reporter.New(logger, config),
		logger:    logger,
		config:    config,
	}
}

// Run function starts agent
func (a *Agent) Run(ctx context.Context) {
	pollTimer := time.NewTicker(a.config.PollInterval)
	defer pollTimer.Stop()
	reportTimer := time.NewTicker(a.config.ReportInterval)
	defer reportTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			a.logger.Info("agent gracefully stopped")
			return
		case <-pollTimer.C:
			a.logger.Debug("time to poll metrics")
			a.metricsToReport = a.poller.Poll()
		case <-reportTimer.C:
			a.logger.Debug("time to report metrics")
			a.reporter.Report(a.generator.Generate(a.metricsToReport))
			a.logger.Debug("generated requests and reported metrics")
		}
	}
}
