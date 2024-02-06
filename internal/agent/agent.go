package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/mem"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
)

type Agent struct {
	client  *http.Client
	params  *Params
	metrics *metric.Metrics
}

type Params struct {
	Address        string
	PollInterval   time.Duration
	ReportInterval time.Duration
	HashKey        string
	RateLimit      int
}

func New(params *Params) *Agent {
	return &Agent{
		client:  &http.Client{},
		params:  params,
		metrics: metric.New(),
	}
}

func (a *Agent) Run(ctx context.Context, workers int) {
	wg := new(sync.WaitGroup)

	wg.Add(2)

	go a.pollMetrics(ctx, wg)
	go a.reportMetrics(ctx, wg, workers)

	<-ctx.Done()

	wg.Wait()

	log.Println("agent gracefully stopped")

}

func (a *Agent) pollMetrics(ctx context.Context, wg *sync.WaitGroup) {
	pollTicker := time.NewTicker(a.params.PollInterval)
	defer pollTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-pollTicker.C:
			a.updateRuntimeMetrics()
			a.updateUtilMetrics()
		}
	}

}

func (a *Agent) updateRuntimeMetrics() {

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	a.metrics.Gauges[metric.Alloc] = metric.Gauge(ms.Alloc)
	a.metrics.Gauges[metric.BuckHashSys] = metric.Gauge(ms.BuckHashSys)
	a.metrics.Gauges[metric.Frees] = metric.Gauge(ms.Frees)
	a.metrics.Gauges[metric.GCCPUFraction] = metric.Gauge(ms.GCCPUFraction)
	a.metrics.Gauges[metric.GCSys] = metric.Gauge(ms.GCSys)
	a.metrics.Gauges[metric.HeapAlloc] = metric.Gauge(ms.HeapAlloc)
	a.metrics.Gauges[metric.HeapIdle] = metric.Gauge(ms.HeapIdle)
	a.metrics.Gauges[metric.HeapInuse] = metric.Gauge(ms.HeapInuse)
	a.metrics.Gauges[metric.HeapObjects] = metric.Gauge(ms.HeapObjects)
	a.metrics.Gauges[metric.HeapReleased] = metric.Gauge(ms.HeapReleased)
	a.metrics.Gauges[metric.HeapSys] = metric.Gauge(ms.HeapSys)
	a.metrics.Gauges[metric.LastGC] = metric.Gauge(ms.LastGC)
	a.metrics.Gauges[metric.Lookups] = metric.Gauge(ms.Lookups)
	a.metrics.Gauges[metric.MCacheInuse] = metric.Gauge(ms.MCacheInuse)
	a.metrics.Gauges[metric.MCacheSys] = metric.Gauge(ms.MCacheSys)
	a.metrics.Gauges[metric.MSpanInuse] = metric.Gauge(ms.MSpanInuse)
	a.metrics.Gauges[metric.MSpanSys] = metric.Gauge(ms.MSpanSys)
	a.metrics.Gauges[metric.Mallocs] = metric.Gauge(ms.Mallocs)
	a.metrics.Gauges[metric.NextGC] = metric.Gauge(ms.NextGC)
	a.metrics.Gauges[metric.NumForcedGC] = metric.Gauge(ms.NumForcedGC)
	a.metrics.Gauges[metric.NumGC] = metric.Gauge(ms.NumGC)
	a.metrics.Gauges[metric.OtherSys] = metric.Gauge(ms.OtherSys)
	a.metrics.Gauges[metric.PauseTotalNs] = metric.Gauge(ms.PauseTotalNs)
	a.metrics.Gauges[metric.StackInuse] = metric.Gauge(ms.StackInuse)
	a.metrics.Gauges[metric.StackSys] = metric.Gauge(ms.StackSys)
	a.metrics.Gauges[metric.Sys] = metric.Gauge(ms.Sys)
	a.metrics.Gauges[metric.TotalAlloc] = metric.Gauge(ms.TotalAlloc)

	a.metrics.Counters[metric.PollCount]++
	a.metrics.Gauges[metric.RandomValue] = metric.Gauge(rand.Float64())
}

func (a *Agent) updateUtilMetrics() {
	v, _ := mem.VirtualMemory()

	a.metrics.Gauges[metric.TotalMemory] = metric.Gauge(v.Total)
	a.metrics.Gauges[metric.FreeMemory] = metric.Gauge(v.Free)

}

func (a *Agent) reportMetrics(ctx context.Context, wg *sync.WaitGroup, workers int) {
	reportTicker := time.NewTicker(a.params.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-reportTicker.C:
			requests := a.requestGenerator(workers)
			for i := 1; i <= workers; i++ {
				go a.requestExecutor(ctx, requests, 3)
			}

		}
	}

}

func (a *Agent) requestGenerator(workers int) chan *http.Request {
	requests := make(chan *http.Request, workers)

	go func() {
		go a.makeBatchMetricsRequest(requests)
		//go a.makeGaugeMetricRequest(requests)
		//go a.makeCounterMetricRequest(requests)
		//go a.makeMetricsGZIPRequest(requests)
	}()

	return requests
}

func (a *Agent) requestExecutor(ctx context.Context, requests chan *http.Request, retryAttempts int) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-requests:
			resp, err := a.client.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()
		}
	}
}

func (a *Agent) makeGaugeMetricRequest(requests chan *http.Request) {
	for name, value := range a.metrics.Gauges {
		stringVal := strconv.FormatFloat(float64(value), 'g', -1, 64)
		urlEndpoint := fmt.Sprintf("http://%s/update/gauge/%s/%s", a.params.Address, string(name), stringVal)
		req, err := http.NewRequest(http.MethodPost, urlEndpoint, nil)
		if err != nil {
			continue
		}
		requests <- req
	}
}

func (a *Agent) makeCounterMetricRequest(requests chan *http.Request) {
	for name, value := range a.metrics.Counters {
		stringVal := strconv.FormatInt(int64(value), 10)
		urlEndpoint := fmt.Sprintf("http://%s/update/counter/%s/%s", a.params.Address, string(name), stringVal)
		req, err := http.NewRequest(http.MethodPost, urlEndpoint, nil)
		if err != nil {
			continue
		}
		a.metrics.Counters[name] = 0
		requests <- req
	}
}

func (a *Agent) makeMetricsGZIPRequest(requests chan *http.Request) {
	for name, value := range a.metrics.Gauges {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "gauge"
		val := float64(value)
		m.Value = &val
		metricJSON, err := json.Marshal(m)
		if err != nil {
			continue
		}
		compressedMetric, err := compress(metricJSON)
		if err != nil {
			continue
		}

		endpoint := fmt.Sprintf("http://%s/update/", a.params.Address)

		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(compressedMetric))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Encoding", "gzip")
		requests <- req
	}
	for name, value := range a.metrics.Counters {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "counter"
		val := int64(value)
		m.Delta = &val
		metricJSON, err := json.Marshal(m)
		if err != nil {
			continue
		}
		compressedMetric, err := compress(metricJSON)
		if err != nil {
			continue
		}

		endpoint := fmt.Sprintf("http://%s/update/", a.params.Address)

		req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(compressedMetric))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Encoding", "gzip")
		requests <- req
	}

}

func (a *Agent) makeBatchMetricsRequest(requests chan *http.Request) {
	var metricsBatch []metric.MetricDTO
	for name, value := range a.metrics.Gauges {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "gauge"
		val := float64(value)
		m.Value = &val

		metricsBatch = append(metricsBatch, m)
	}

	for name, value := range a.metrics.Counters {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "counter"
		val := int64(value)
		m.Delta = &val

		metricsBatch = append(metricsBatch, m)

		a.metrics.Counters[name] = 0
	}

	metricsBatchJSON, _ := json.Marshal(metricsBatch)

	compressedMetricsBatch, _ := compress(metricsBatchJSON)

	endpoint := fmt.Sprintf("http://%s/updates/", a.params.Address)

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(compressedMetricsBatch))

	req.Header.Set("Content-Encoding", "gzip")

	if a.params.HashKey != "" {
		var data []byte
		req.Body.Read(data)

		h := hmac.New(sha256.New, []byte(a.params.HashKey))
		h.Write(data)
		sign := h.Sum(nil)
		signBase64 := base64.StdEncoding.EncodeToString(sign)

		req.Header.Add("HashSHA256", signBase64)
	}
	requests <- req
}

func (a *Agent) reportGaugeMetric(name metric.Name, value metric.Gauge) error {
	const op = "agent.reportGaugeMetric()"
	stringVal := strconv.FormatFloat(float64(value), 'g', -1, 64)

	endpoint := fmt.Sprintf("http://%s/update/gauge/%s/%s", a.params.Address, string(name), stringVal)

	resp, err := a.client.Post(endpoint, "text/plain", nil)
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)

	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "failed to report metric", err)
	}
	return nil
}

func (a *Agent) reportCounterMetric(name metric.Name, value metric.Counter) error {
	const op = "agent.reportCounterMetric()"

	stringVal := strconv.FormatInt(int64(value), 10)

	endpoint := fmt.Sprintf("http://%s/update/counter/%s/%s", a.params.Address, string(name), stringVal)

	resp, err := a.client.Post(endpoint, "text/plain", nil)
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "failed to report metric", err)
	}
	a.metrics.Counters[name] = 0
	return nil
}

func (a *Agent) reportGaugeMetricJSON(name metric.Name, value metric.Gauge) error {
	const op = "agent.reportGaugeMetricJSON()"

	var m metric.MetricDTO

	m.ID = string(name)
	m.MType = "gauge"
	val := float64(value)
	m.Value = &val

	metricJSON, err := json.Marshal(m)
	if err != nil {
		return e.WrapError(op, "failed to marshal metric", err)
	}

	endpoint := fmt.Sprintf("http://%s/update/", a.params.Address)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(metricJSON))
	if err != nil {
		return e.WrapError(op, "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "status code is not OK", nil)
	}
	return nil

}

func (a *Agent) reportCounterMetricJSON(name metric.Name, value metric.Counter) error {
	const op = "agent.reportCounterMetricJSON()"

	var m metric.MetricDTO

	m.ID = string(name)
	m.MType = "counter"
	val := int64(value)
	m.Delta = &val

	metricJSON, err := json.Marshal(m)
	if err != nil {
		return e.WrapError(op, "failed to marshal metric", err)
	}

	endpoint := fmt.Sprintf("http://%s/update/", a.params.Address)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(metricJSON))
	if err != nil {
		return e.WrapError(op, "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "status code is not OK", err)
	}

	a.metrics.Counters[name] = 0
	return nil
}

func (a *Agent) reportGaugeMetricGZIP(name metric.Name, value metric.Gauge) error {
	const op = "agent.reportGaugeMetricGZIP()"

	var m metric.MetricDTO

	m.ID = string(name)
	m.MType = "gauge"
	val := float64(value)
	m.Value = &val

	metricJSON, err := json.Marshal(m)
	if err != nil {
		return e.WrapError(op, "failed to marshal metric", err)
	}

	compressedMetricJSON, err := compress(metricJSON)
	if err != nil {
		return e.WrapError(op, "failed to compress metric", err)
	}

	endpoint := fmt.Sprintf("http://%s/update/", a.params.Address)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(compressedMetricJSON))
	if err != nil {
		return e.WrapError(op, "failed to create request", err)
	}
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := a.client.Do(req)
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "status code is not OK", nil)
	}
	return nil

}

func (a *Agent) reportCounterMetricGZIP(name metric.Name, value metric.Counter) error {
	const op = "agent.reportCounterMetricGZIP()"

	var m metric.MetricDTO

	m.ID = string(name)
	m.MType = "counter"
	val := int64(value)
	m.Delta = &val

	metricJSON, err := json.Marshal(m)
	if err != nil {
		return e.WrapError(op, "failed to marshal metric", err)
	}

	compressedMetric, err := compress(metricJSON)
	if err != nil {
		return e.WrapError(op, "failed to compress metric", err)
	}

	endpoint := fmt.Sprintf("http://%s/update/", a.params.Address)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(compressedMetric))
	if err != nil {
		return e.WrapError(op, "failed to create request", err)
	}
	req.Header.Set("Content-Encoding", "gzip")

	resp, err := a.client.Do(req)
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "status code is not OK", err)
	}

	a.metrics.Counters[name] = 0
	return nil
}

func (a *Agent) reportMetricsBatch() error {
	const op = "agent.reportMetricsBatch()"

	var metricsBatch []metric.MetricDTO
	for name, value := range a.metrics.Gauges {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "gauge"
		val := float64(value)
		m.Value = &val

		metricsBatch = append(metricsBatch, m)
	}

	for name, value := range a.metrics.Counters {
		var m metric.MetricDTO
		m.ID = string(name)
		m.MType = "counter"
		val := int64(value)
		m.Delta = &val

		metricsBatch = append(metricsBatch, m)

		a.metrics.Counters[name] = 0
	}

	metricsBatchJSON, err := json.Marshal(metricsBatch)
	if err != nil {
		return e.WrapError(op, "failed to marshal metrics", err)
	}

	compressedMetricsBatch, err := compress(metricsBatchJSON)
	if err != nil {
		return e.WrapError(op, "failed to compress metrics", err)
	}

	endpoint := fmt.Sprintf("http://%s/updates/", a.params.Address)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(compressedMetricsBatch))
	if err != nil {
		return e.WrapError(op, "failed to create request", err)
	}
	req.Header.Set("Content-Encoding", "gzip")

	if a.params.HashKey != "" {
		var data []byte
		_, err := req.Body.Read(data)
		if err != nil {
			return e.WrapError(op, "can not read body", err)
		}

		h := hmac.New(sha256.New, []byte(a.params.HashKey))
		h.Write(data)
		sign := h.Sum(nil)
		signBase64 := base64.StdEncoding.EncodeToString(sign)

		req.Header.Add("HashSHA256", signBase64)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return e.WrapError(op, "failed to report metrics", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "status code is not OK", err)
	}

	return nil
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

func retry(attempts int, f func() error) (err error) {
	sleep := 1 * time.Second
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep += 2 * time.Second
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
