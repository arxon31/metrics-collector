package generator

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	config "github.com/arxon31/metrics-collector/internal/config/agent"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
)

type requestGenerator struct {
	config            *config.Config
	logger            *zap.SugaredLogger
	rw                sync.RWMutex
	metrics           *metric.Metrics
	generatedRequests chan *http.Request
	errChan           chan error
	res               *result
}

type result struct {
	errors int64
	all    int64
}

func (r *result) incrementAll() {
	atomic.AddInt64(&r.all, 1)
}

func (r *result) incrementError() {
	atomic.AddInt64(&r.errors, 1)
}

func (r *result) reset() {
	r.errors = 0
	r.all = 0
}

func New(config *config.Config, logger *zap.SugaredLogger) *requestGenerator {
	generated := make(chan *http.Request, config.RateLimit)
	g := &requestGenerator{
		config:            config,
		generatedRequests: generated,
		errChan:           make(chan error),
		logger:            logger,
		res: &result{
			errors: 0,
			all:    0,
		},
	}

	return g
}

func (g *requestGenerator) Generate(metrics *metric.Metrics) chan *http.Request {
	g.metrics = metrics
	done := make(chan struct{})

	go g.errorLogger(done)

	go g.makeBatchMetricsRequest(done)

	go g.makeReport(done)

	return g.generatedRequests
}

func (g *requestGenerator) errorLogger(done chan struct{}) {
	for {
		select {
		case err := <-g.errChan:
			g.logger.Info(err)
		case <-done:
			return
		}
	}
}

func (g *requestGenerator) makeReport(done chan struct{}) {
	<-done
	g.logger.Infof("From making %d requests have %d errors", g.res.all, g.res.errors)
	g.res.reset()

}

func (g *requestGenerator) makeGaugeMetricRequest(done chan struct{}) {
	const op = "request-generator.makeGaugeMetricRequest()"

	defer close(done)

	g.rw.RLock()
	defer g.rw.RUnlock()
	for name, value := range g.metrics.Gauges {
		g.res.incrementAll()
		stringVal := strconv.FormatFloat(float64(value), 'g', -1, 64)
		urlEndpoint := fmt.Sprintf("http://%s/update/gauge/%s/%s", g.config.Address, string(name), stringVal)
		req, err := http.NewRequest(http.MethodPost, urlEndpoint, nil)
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not make URL request "+urlEndpoint, err)
			continue
		}
		g.generatedRequests <- req
	}
}

func (g *requestGenerator) makeCounterMetricRequest(done chan struct{}) {
	const op = "request-generator.makeGaugeMetricRequest()"

	defer close(done)

	g.rw.Lock()
	defer g.rw.Unlock()

	for name, value := range g.metrics.Counters {
		g.res.incrementAll()
		stringVal := strconv.FormatInt(int64(value), 10)
		urlEndpoint := fmt.Sprintf("http://%s/update/counter/%s/%s", g.config.Address, string(name), stringVal)
		req, err := http.NewRequest(http.MethodPost, urlEndpoint, nil)
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not make URL request: "+urlEndpoint, err)
			continue
		}

		g.metrics.Counters[name] = 0

		g.generatedRequests <- req
	}
}

func (g *requestGenerator) makeMetricsGZIPRequest(done chan struct{}) {
	const op = "request-generator.makeMetricsGZIPRequest()"

	defer close(done)

	endpoint := fmt.Sprintf("http://%s/update/", g.config.Address)
	g.rw.RLock()
	for name, value := range g.metrics.Gauges {
		g.res.incrementAll()
		val := float64(value)
		m := metric.MetricDTO{
			ID:    string(name),
			MType: "gauge",
			Value: &val,
		}

		metricJSON, err := json.Marshal(m)
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not marshal metric: "+m.ID, err)
			continue
		}

		compressedMetric, err := compress(metricJSON)
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not compress metric: "+m.ID, err)
			continue
		}

		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(compressedMetric))
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not make GZIP request for metric: "+m.ID, err)
			continue
		}
		req.Header.Set("Content-Encoding", "gzip")

		g.generatedRequests <- req
	}
	g.rw.RUnlock()

	g.rw.Lock()
	for name, value := range g.metrics.Counters {
		g.res.incrementAll()
		val := int64(value)
		m := metric.MetricDTO{
			ID:    string(name),
			MType: "counter",
			Delta: &val,
		}

		metricJSON, err := json.Marshal(m)
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not marshal metric: "+m.ID, err)
			continue
		}
		compressedMetric, err := compress(metricJSON)
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not compress metric: "+m.ID, err)
			continue
		}

		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(compressedMetric))
		if err != nil {
			g.res.incrementError()
			g.errChan <- e.WrapError(op, "can not make GZIP request for metric: "+m.ID, err)
			continue
		}
		req.Header.Set("Content-Encoding", "gzip")

		g.metrics.Counters[name] = 0

		g.generatedRequests <- req
	}
	g.rw.Unlock()

}

func (g *requestGenerator) makeBatchMetricsRequest(done chan struct{}) {
	const op = "request-generator.makeBatchMetricsRequest()"
	g.res.incrementAll()

	defer close(done)

	var metricsBatch []metric.MetricDTO
	g.rw.RLock()
	for name, value := range g.metrics.Gauges {
		val := float64(value)
		m := metric.MetricDTO{
			ID:    string(name),
			MType: "gauge",
			Value: &val,
		}

		metricsBatch = append(metricsBatch, m)
	}
	g.rw.RUnlock()

	g.rw.Lock()
	for name, value := range g.metrics.Counters {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "counter"
		val := int64(value)
		m.Delta = &val

		metricsBatch = append(metricsBatch, m)

		g.metrics.Counters[name] = 0
	}
	g.rw.Unlock()

	metricsBatchJSON, err := json.Marshal(metricsBatch)
	if err != nil {
		g.res.incrementError()
		g.errChan <- e.WrapError(op, "can not marshal metrics batch", err)
		return

	}

	compressedMetricsBatch, err := compress(metricsBatchJSON)
	if err != nil {
		g.res.incrementError()
		g.errChan <- e.WrapError(op, "can not compress metrics batch", err)
		return
	}

	endpoint := fmt.Sprintf("http://%s/updates/", g.config.Address)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(compressedMetricsBatch))
	if err != nil {
		g.res.incrementError()
		g.errChan <- e.WrapError(op, "can not make request with metrics batch", err)
		return
	}

	req.Header.Set("Content-Encoding", "gzip")

	if g.config.HashKey != "" {
		var data []byte
		req.Body.Read(data)

		sign := g.countHash(data)

		req.Header.Add("HashSHA256", sign)
	}
	g.generatedRequests <- req
}

func compress(b []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(buf)

	_, err := gz.Write(b)
	if err != nil {
		return nil, err
	}
	gz.Close()

	return buf.Bytes(), nil
}

func (g *requestGenerator) countHash(data []byte) (sign string) {
	h := hmac.New(sha256.New, []byte(g.config.HashKey))
	h.Write(data)
	s := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(s)

}
