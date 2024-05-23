// Package reporter receives http requests from generator and report them to server
package reporter

import (
	"context"
	"errors"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"time"

	"go.uber.org/zap"
)

const retryAttempts = 3
const workerTimeout = 2 * time.Second

type reporter interface {
	DoCtx(ctx context.Context, req *resty.Request) (*resty.Response, error)
}

type metricReporter struct {
	logger         *zap.SugaredLogger
	rateLimit      int
	reportInterval time.Duration
	reporter       reporter
	requests       <-chan *resty.Request
	workers        errgroup.Group
}

func NewReporter(logger *zap.SugaredLogger, rateLimit int, reportInterval time.Duration, reporter reporter, requests <-chan *resty.Request) *metricReporter {
	workers := errgroup.Group{}
	workers.SetLimit(rateLimit)

	rep := &metricReporter{
		logger:         logger,
		reportInterval: reportInterval,
		reporter:       reporter,
		requests:       requests,
		workers:        workers,
	}

	return rep
}

func (r *metricReporter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.reportInterval)
	defer ticker.Stop()

	workerCtx, cancel := context.WithTimeout(ctx, workerTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			err := r.workers.Wait()
			if err != nil && !errors.Is(err, context.Canceled) {
				r.logger.Error(err)
			}
			r.logger.Info("reporter gracefully stopped")
			return
		case <-ticker.C:
			for r.workers.TryGo(func() error {
				return r.runWorker(workerCtx)
			}) {
			}

		}

	}
}

func (r *metricReporter) runWorker(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case request, ok := <-r.requests:
			if !ok {
				return nil
			}
			resp, err := r.reporter.DoCtx(ctx, request)
			if err != nil {
				return err
			}
			if !resp.IsSuccess() {
				r.logger.Warn("reporter failed to send request", zap.String("request", request.URL), zap.Int("status code", resp.StatusCode()))
			}
		}
	}
}
