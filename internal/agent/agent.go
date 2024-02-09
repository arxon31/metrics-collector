package agent

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/agent/generator"
	rgen "github.com/arxon31/metrics-collector/internal/agent/generator/request-generator"
	"github.com/arxon31/metrics-collector/internal/agent/poller"
	mpoll "github.com/arxon31/metrics-collector/internal/agent/poller/metric-poller"
	"github.com/arxon31/metrics-collector/internal/agent/reporter"
	mrep "github.com/arxon31/metrics-collector/internal/agent/reporter/metric-reporter"
	config "github.com/arxon31/metrics-collector/internal/config/agent"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"go.uber.org/zap"
	"time"
)

type Agent struct {
	poller          poller.Poller
	generator       generator.Generator
	reporter        reporter.Reporter
	logger          *zap.SugaredLogger
	config          *config.Config
	metricsToReport *metric.Metrics
}

func New(config *config.Config, logger *zap.SugaredLogger) *Agent {
	p := mpoll.New(logger)
	g := rgen.New(config, logger)
	r := mrep.New(logger, config)
	return &Agent{
		poller:    p,
		generator: g,
		reporter:  r,
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
