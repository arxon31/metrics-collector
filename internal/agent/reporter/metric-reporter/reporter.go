package reporter

import (
	"fmt"
	"github.com/arxon31/metrics-collector/internal/agent/reporter"
	config "github.com/arxon31/metrics-collector/internal/config/agent"
	"github.com/arxon31/metrics-collector/pkg/e"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

const retryAttempts = 3

type metricReporter struct {
	logger  *zap.SugaredLogger
	config  *config.Config
	errChan chan error
	result  *result
	client  *http.Client
}

type result struct {
	all, errors int
}

func (r *result) incrementAll() {
	r.all++
}
func (r *result) incrementError() {
	r.errors++
}
func (r *result) reset() {
	r.all = 0
	r.errors = 0
}

func (r *metricReporter) makeReport() {
	r.logger.Infof("From executing %d requests have %d errors", r.result.all, r.result.errors)
	r.result.reset()
}

func New(logger *zap.SugaredLogger, config *config.Config) reporter.Reporter {
	return &metricReporter{
		logger:  logger,
		config:  config,
		errChan: make(chan error),
		client:  &http.Client{},
		result: &result{
			all:    0,
			errors: 0,
		},
	}
}

func (r *metricReporter) Report(requests chan *http.Request) {
	var wg sync.WaitGroup

	done := make(chan struct{})

	go r.errorLogger(done)

	for i := 1; i <= r.config.RateLimit; i++ {
		go r.requestExecutor(requests, &wg)
	}
	wg.Wait()

	done <- struct{}{}

}

func (r *metricReporter) errorLogger(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case err := <-r.errChan:
			r.logger.Info(err)
		}
	}
}

func (r *metricReporter) requestExecutor(requests chan *http.Request, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	r.executeRequest(<-requests)

}

func (r *metricReporter) executeRequest(req *http.Request) {
	const op = "metric-reporter.executeRequest()"
	r.result.incrementAll()
	r.logger.Debug("executing request " + req.URL.String())
	resp, err := r.client.Do(req)
	if err != nil {
		r.result.incrementError()
		r.errChan <- e.WrapError(op, "can not execute request. Retrying", err)
		err = r.retry(r.client.Do, req)
		if err != nil {
			r.errChan <- err
			return
		}

	}
	defer resp.Body.Close()
	r.logger.Debug("request " + req.URL.String() + " successfully executed")
}

func (r *metricReporter) retry(f func(req *http.Request) (*http.Response, error), req *http.Request) (e error) {
	sleep := 1 * time.Second
	var err error
	for i := 0; i < retryAttempts; i++ {
		resp, err := f(req)
		if err == nil {
			resp.Body.Close()
			r.logger.Infof("successful execution after %d retries", i)
			return nil
		}
		r.result.incrementError()
		time.Sleep(sleep)
		sleep += 2 * time.Second
	}
	return fmt.Errorf("after %d attempts, last error: %w", retryAttempts, err)
}
