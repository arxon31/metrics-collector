package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
}

func New(params *Params) *Agent {
	return &Agent{
		client:  &http.Client{},
		params:  params,
		metrics: metric.New(),
	}
}

func (a *Agent) Run(ctx context.Context) {
	wg := new(sync.WaitGroup)

	wg.Add(2)

	go a.pollMetrics(ctx, wg)
	go a.reportMetrics(ctx, wg)

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
			a.update()
		}
	}

}

func (a *Agent) update() {

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

func (a *Agent) reportMetrics(ctx context.Context, wg *sync.WaitGroup) {
	reportTicker := time.NewTicker(a.params.ReportInterval)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-reportTicker.C:
			for k, val := range a.metrics.Gauges {
				if err := a.reportGaugeMetricJSON(k, val); err != nil {
					log.Println(err)
					continue
				}
				log.Printf("reported metric %s with value %f\n", k, val)
			}
			for k, val := range a.metrics.Counters {
				if err := a.reportCounterMetricJSON(k, val); err != nil {
					log.Println(err)
					continue
				}
				log.Printf("reported metric %s with value %v\n", k, val)
			}
		}
	}

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

	resp, err := a.client.Post(endpoint, "application/json", bytes.NewBuffer(metricJSON))
	if err != nil {
		return e.WrapError(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.WrapError(op, "failed to report metric", err)
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

	resp, err := a.client.Post(endpoint, "application/json", bytes.NewBuffer(metricJSON))
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
