// Package reporter receives http requests from generator and report them to server
package reporter

import (
	"context"
	"net/http"
	"time"

	"github.com/arxon31/metrics-collector/pkg/logger"

	"go.uber.org/zap"
)

const workerTimeout = 2 * time.Second

type reporter interface {
	Do(req *http.Request) (*http.Response, error)
}

type metricReporter struct {
	rateLimit      int
	reportInterval time.Duration
	reporter       reporter
}

// NewReporter creates new reporter
func NewReporter(rateLimit int, reporter reporter) *metricReporter {

	rep := &metricReporter{
		reporter:  reporter,
		rateLimit: rateLimit,
	}

	return rep
}

// Report func sends rest requests to server
func (r *metricReporter) Report(reqChan <-chan *http.Request) {
	for i := 0; i < r.rateLimit; i++ {
		go r.runWorker(reqChan)
	}
}

func (r *metricReporter) runWorker(reqChan <-chan *http.Request) error {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), workerTimeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil
		case req, ok := <-reqChan:
			if !ok {
				return nil
			}
			resp, err := r.reporter.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				logger.Logger.Error("unexpected status code", zap.Int("status_code", resp.StatusCode))
			}
			logger.Logger.Info("request processed")
			return nil
		}
	}

}
