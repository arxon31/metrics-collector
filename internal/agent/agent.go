package agent

import (
	"context"
	"fmt"
	"github.com/arxon31/metrics-collector/pkg/e"
	"github.com/arxon31/metrics-collector/pkg/metric"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

type Agent struct {
	client  *http.Client
	params  *Params
	metrics *metric.Metrics
	ctx     context.Context
}

type Params struct {
	Address        string
	PollInterval   int
	ReportInterval int
}

func New(ctx context.Context, params *Params) *Agent {
	metrics := metric.New()
	return &Agent{
		client:  &http.Client{},
		params:  params,
		metrics: metrics,
		ctx:     ctx,
	}
}

func (a *Agent) Run() {
	go a.pollMetrics()
	a.reportMetrics()
}

func (a *Agent) pollMetrics() {
	for range time.Tick(time.Duration(a.params.PollInterval) * time.Second) {
		a.update()
	}
}

func (a *Agent) update() {
	a.metrics.RW.Lock()
	defer a.metrics.RW.Unlock()

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

	a.metrics.Counters[metric.PollCount] += 1
	a.metrics.Gauges[metric.RandomValue] = metric.Gauge(rand.Float64())
}

func (a *Agent) reportMetrics() {
	for range time.Tick(time.Duration(a.params.ReportInterval) * time.Second) {
		for k, val := range a.metrics.Gauges {
			if err := a.reportGaugeMetric(k, val); err != nil {
				fmt.Println(err)
				continue
			}
			log.Printf("reported metric %s with value %f\n", k, val)
		}
		for k, val := range a.metrics.Counters {
			if err := a.reportCounterMetric(k, val); err != nil {
				fmt.Println(err)
				continue
			}
			log.Printf("reported metric %s with value %v\n", k, val)
		}
	}

}

func (a *Agent) reportGaugeMetric(name metric.Name, value metric.Gauge) error {
	const op = "agent.reportGaugeMetric()"
	stringVal := strconv.FormatFloat(float64(value), 'g', -1, 64)

	endpoint := fmt.Sprintf("http://%s/update/gauge/%s/%s", a.params.Address, string(name), stringVal)

	resp, err := a.client.Post(endpoint, "text/plain", nil)
	if err != nil {
		return e.Wrap(op, "failed to report metric", err)

	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.Wrap(op, "failed to report metric", err)
	}
	return nil
}

func (a *Agent) reportCounterMetric(name metric.Name, value metric.Counter) error {
	const op = "agent.reportCounterMetric()"

	stringVal := strconv.FormatInt(int64(value), 10)

	endpoint := fmt.Sprintf("http://%s/update/counter/%s/%s", a.params.Address, string(name), stringVal)

	resp, err := a.client.Post(endpoint, "text/plain", nil)
	if err != nil {
		return e.Wrap(op, "failed to report metric", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e.Wrap(op, "failed to report metric", err)
	}
	a.metrics.Counters[name] = 0
	return nil
}
