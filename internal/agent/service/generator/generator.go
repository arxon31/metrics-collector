// Package generator generates http requests for each metrics update and sends requests to reporter
package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

const (
	metricURL = "update"
	batchURL  = "updates"
)

const hashHeader = "HashSHA256"

type repo interface {
	StoreCounter(ctx context.Context, name string, value int64) error
	Metrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type hasher interface {
	Hash([]byte) (string, error)
}

type compressor interface {
	Compress([]byte) ([]byte, error)
}

type requestGenerator struct {
	address           string
	rateLimit         int
	repo              repo
	hasher            hasher
	compressor        compressor
	logger            *zap.SugaredLogger
	GeneratedRequests chan *resty.Request
}

func New(address string, rateLimit int, repo repo, hasher hasher, compressor compressor, logger *zap.SugaredLogger) *requestGenerator {
	g := &requestGenerator{
		address:           address,
		rateLimit:         rateLimit,
		repo:              repo,
		hasher:            hasher,
		compressor:        compressor,
		logger:            logger,
		GeneratedRequests: make(chan *resty.Request, rateLimit),
	}

	return g
}

// Generate func generating requests and sending them to generated channel
// Below you can see all the methods that can be used
// Now using only makeBatchMetricsRequest which generates request with all metrics in JSON
func (g *requestGenerator) Generate(ctx context.Context) {
	go g.makeBatchMetricsRequest(ctx)
}

func (g *requestGenerator) Requests() <-chan *resty.Request {
	return g.GeneratedRequests
}

func (g *requestGenerator) makeGaugeURLRequest(ctx context.Context) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		g.logger.Error(err)
		return
	}

	for _, metric := range metrics {
		if metric.MetricType == entity.GaugeType {
			val := strconv.FormatFloat(*metric.Gauge, 'f', -1, 64)
			url := g.makeUrl(metric.MetricType, metric.Name, val)
			req := resty.Request{URL: url, Method: http.MethodPost}
			g.GeneratedRequests <- &req
		}
	}

}

func (g *requestGenerator) makeCounterURLRequest(ctx context.Context) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		g.logger.Error(err)
		return
	}

	for _, metric := range metrics {
		if metric.MetricType == entity.CounterType {
			val := strconv.FormatInt(*metric.Counter, 10)
			url := g.makeUrl(metric.MetricType, metric.Name, val)
			req := resty.Request{URL: url, Method: http.MethodPost}
			g.GeneratedRequests <- &req
		}
	}
}

func (g *requestGenerator) makeCompressedMetricsRequest(ctx context.Context) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		g.logger.Error(err)
		return
	}

	url := g.makeUrl2(metricURL)

	for _, metric := range metrics {
		metricBytes, err := json.Marshal(metric)
		if err != nil {
			g.logger.Error(err)
			continue
		}
		compressedMetric, err := g.compressor.Compress(metricBytes)
		if err != nil {
			g.logger.Error(err)
			continue
		}
		req := resty.Request{URL: url, Method: http.MethodPost, Body: compressedMetric}
		req.Header.Set("Content-Encoding", "gzip")
		g.GeneratedRequests <- &req

		if metric.MetricType == entity.CounterType {
			err := g.repo.StoreCounter(ctx, metric.Name, -*metric.Counter)
			if err != nil {
				g.logger.Error(err)
			}
		}

	}
}

func (g *requestGenerator) makeBatchMetricsRequest(ctx context.Context) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		g.logger.Error(err)
		return
	}

	metricsBatchJSON, err := json.Marshal(metrics)
	if err != nil {
		g.logger.Error(err)
		return
	}

	metricsBatchCompressed, err := g.compressor.Compress(metricsBatchJSON)
	if err != nil {
		g.logger.Error(err)
		return
	}
	url := g.makeUrl2(batchURL)
	req := resty.Request{URL: url, Method: http.MethodPost, Body: metricsBatchCompressed, Header: http.Header{"Content-Encoding": {"gzip"}}}

	hashSign, err := g.hasher.Hash(metricsBatchCompressed)
	if err != nil {
		g.GeneratedRequests <- &req
		return
	}
	req.Header.Set(hashHeader, hashSign)
	g.GeneratedRequests <- &req
}

func (g *requestGenerator) makeUrl(metricType, name, val string) string {
	return fmt.Sprintf("http://%s/update/%s/%s/%s", g.address, metricType, name, val)
}

func (g *requestGenerator) makeUrl2(endpoint string) string {
	return fmt.Sprintf("http://%s/%s/", g.address, endpoint)
}
