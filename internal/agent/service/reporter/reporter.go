// Package reporter receives http requests from generator and report them to server
package reporter

import (
	"context"
	"github.com/go-resty/resty/v2"
	"time"

	"go.uber.org/zap"
)

const retryAttempts = 3

type reporter interface {
	DoCtx(ctx context.Context, req *resty.Request) (*resty.Response, error)
}

type metricReporter struct {
	logger         *zap.SugaredLogger
	rateLimit      int
	reportInterval time.Duration
	reporter       reporter
	requests       chan *resty.Request
}

func New(logger *zap.SugaredLogger, rateLimit int, reportInterval time.Duration, reporter reporter, requests chan *resty.Request) *metricReporter {
	return &metricReporter{
		logger:         logger,
		rateLimit:      rateLimit,
		reportInterval: reportInterval,
		reporter:       reporter,
		requests:       requests,
	}
}

// Run function asynchronous sending requests to server
func (r *metricReporter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("reporter gracefully stopped")
			return
		case <-ticker.C:
			req := <-r.requests
			if req != nil {
				go r.reporter.DoCtx(ctx, req)
			}
		}
	}
}
