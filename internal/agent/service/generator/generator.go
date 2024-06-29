// Package generator generates http requests for each metrics update and sends requests to reporter
package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/pkg/logger"
	"net/http"
	"strconv"

	"github.com/arxon31/metrics-collector/internal/entity"
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
	address    string
	rateLimit  int
	repo       repo
	hasher     hasher
	compressor compressor
}

func New(address string, repo repo, hasher hasher, compressor compressor) *requestGenerator {
	g := &requestGenerator{
		address:    address,
		repo:       repo,
		hasher:     hasher,
		compressor: compressor,
	}

	return g
}

// Generate func generating requests and sending them to generated channel
// Below you can see all the methods that can be used
func (g *requestGenerator) Generate(ctx context.Context) <-chan *http.Request {
	requests := make(chan *http.Request)

	go g.makeBatchMetricsRequest(ctx, requests)
	go g.makeCompressedMetricsRequest(ctx, requests)

	return requests
}

func (g *requestGenerator) makeGaugeURLRequest(ctx context.Context, requests chan *http.Request) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	for _, metric := range metrics {
		if metric.MetricType == entity.GaugeType {
			val := strconv.FormatFloat(*metric.Gauge, 'f', -1, 64)
			url := g.makeURL(metric.MetricType, metric.Name, val)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				logger.Logger.Error(err)
				continue
			}
			requests <- req
		}
	}

}

func (g *requestGenerator) makeCounterURLRequest(ctx context.Context, requests chan *http.Request) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	for _, metric := range metrics {
		if metric.MetricType == entity.CounterType {
			val := strconv.FormatInt(*metric.Counter, 10)
			url := g.makeURL(metric.MetricType, metric.Name, val)
			req, err := http.NewRequest(http.MethodPost, url, nil)
			if err != nil {
				logger.Logger.Error(err)
				continue
			}
			requests <- req
		}
	}
}

func (g *requestGenerator) makeCompressedMetricsRequest(ctx context.Context, requests chan *http.Request) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	url := g.makeURL2(metricURL)

	for _, metric := range metrics {
		metricBytes, err := json.Marshal(metric)
		if err != nil {
			logger.Logger.Error(err)
			continue
		}
		compressedMetric, err := g.compressor.Compress(metricBytes)
		if err != nil {
			logger.Logger.Error(err)
			continue
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedMetric))
		if err != nil {
			logger.Logger.Error(err)
			continue
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		requests <- req

		if metric.MetricType == entity.CounterType {
			err := g.repo.StoreCounter(ctx, metric.Name, -*metric.Counter)
			if err != nil {
				logger.Logger.Error(err)
			}
		}

	}
}

func (g *requestGenerator) makeGZIPMetricsRequest(ctx context.Context, requests chan *http.Request) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	metricsBatchJSON, err := json.Marshal(metrics)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	metricsBatchCompressed, err := g.compressor.Compress(metricsBatchJSON)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	url := g.makeURL2(batchURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metricsBatchCompressed))
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	hashSign, err := g.hasher.Hash(metricsBatchCompressed)

	if err != nil {
		requests <- req
		return
	}
	req.Header.Set(hashHeader, hashSign)
	requests <- req
}

func (g *requestGenerator) makeBatchMetricsRequest(ctx context.Context, requests chan *http.Request) {
	metrics, err := g.repo.Metrics(ctx)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	metricsBatchJSON, err := json.Marshal(metrics)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	metricsBatchCompressed, err := g.compressor.Compress(metricsBatchJSON)
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	url := g.makeURL2(batchURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(metricsBatchCompressed))
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	hashSign, err := g.hasher.Hash(metricsBatchCompressed)

	if err != nil {
		requests <- req
		return
	}
	req.Header.Set(hashHeader, hashSign)
	requests <- req
}

func (g *requestGenerator) makeURL(metricType, name, val string) string {
	return fmt.Sprintf("http://%s/update/%s/%s/%s", g.address, metricType, name, val)
}

func (g *requestGenerator) makeURL2(endpoint string) string {
	return fmt.Sprintf("http://%s/%s/", g.address, endpoint)
}
