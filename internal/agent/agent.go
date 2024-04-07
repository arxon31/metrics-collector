package agent

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/internal/agent/generator"
	"github.com/arxon31/metrics-collector/internal/agent/poller"
	"github.com/arxon31/metrics-collector/internal/agent/reporter"
	config "github.com/arxon31/metrics-collector/internal/config/agent"
	"github.com/arxon31/metrics-collector/pkg/metric"
)

type MetricsPoller interface {
	Poll() *metric.Metrics
}

type RequestGenerator interface {
	Generate(*metric.Metrics) chan *http.Request
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
	metricsToReport *metric.Metrics
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
