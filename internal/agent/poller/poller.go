package poller

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"

	"github.com/arxon31/metrics-collector/pkg/metric"
)

type metricPoller struct {
	mu      sync.Mutex
	metrics *metric.Metrics
	logger  *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger) *metricPoller {

	p := &metricPoller{
		metrics: metric.New(),
		logger:  logger,
	}

	return p

}

func (p *metricPoller) Poll() *metric.Metrics {
	p.logger.Debug("start polling metrics")
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go p.updateRuntimeMetrics(wg)
	go p.updateUtilMetrics(wg)

	wg.Wait()
	p.logger.Debug("finished polling metrics")
	return p.metrics

}

func (p *metricPoller) updateRuntimeMetrics(group *sync.WaitGroup) {
	defer group.Done()
	p.logger.Debug("start update runtime metrics")

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics.Gauges[metric.Alloc] = metric.Gauge(ms.Alloc)
	p.metrics.Gauges[metric.BuckHashSys] = metric.Gauge(ms.BuckHashSys)
	p.metrics.Gauges[metric.Frees] = metric.Gauge(ms.Frees)
	p.metrics.Gauges[metric.GCCPUFraction] = metric.Gauge(ms.GCCPUFraction)
	p.metrics.Gauges[metric.GCSys] = metric.Gauge(ms.GCSys)
	p.metrics.Gauges[metric.HeapAlloc] = metric.Gauge(ms.HeapAlloc)
	p.metrics.Gauges[metric.HeapIdle] = metric.Gauge(ms.HeapIdle)
	p.metrics.Gauges[metric.HeapInuse] = metric.Gauge(ms.HeapInuse)
	p.metrics.Gauges[metric.HeapObjects] = metric.Gauge(ms.HeapObjects)
	p.metrics.Gauges[metric.HeapReleased] = metric.Gauge(ms.HeapReleased)
	p.metrics.Gauges[metric.HeapSys] = metric.Gauge(ms.HeapSys)
	p.metrics.Gauges[metric.LastGC] = metric.Gauge(ms.LastGC)
	p.metrics.Gauges[metric.Lookups] = metric.Gauge(ms.Lookups)
	p.metrics.Gauges[metric.MCacheInuse] = metric.Gauge(ms.MCacheInuse)
	p.metrics.Gauges[metric.MCacheSys] = metric.Gauge(ms.MCacheSys)
	p.metrics.Gauges[metric.MSpanInuse] = metric.Gauge(ms.MSpanInuse)
	p.metrics.Gauges[metric.MSpanSys] = metric.Gauge(ms.MSpanSys)
	p.metrics.Gauges[metric.Mallocs] = metric.Gauge(ms.Mallocs)
	p.metrics.Gauges[metric.NextGC] = metric.Gauge(ms.NextGC)
	p.metrics.Gauges[metric.NumForcedGC] = metric.Gauge(ms.NumForcedGC)
	p.metrics.Gauges[metric.NumGC] = metric.Gauge(ms.NumGC)
	p.metrics.Gauges[metric.OtherSys] = metric.Gauge(ms.OtherSys)
	p.metrics.Gauges[metric.PauseTotalNs] = metric.Gauge(ms.PauseTotalNs)
	p.metrics.Gauges[metric.StackInuse] = metric.Gauge(ms.StackInuse)
	p.metrics.Gauges[metric.StackSys] = metric.Gauge(ms.StackSys)
	p.metrics.Gauges[metric.Sys] = metric.Gauge(ms.Sys)
	p.metrics.Gauges[metric.TotalAlloc] = metric.Gauge(ms.TotalAlloc)

	p.metrics.Counters[metric.PollCount]++
	p.metrics.Gauges[metric.RandomValue] = metric.Gauge(rand.Float64())

	p.logger.Debug("successfully updated runtime metrics")
}

func (p *metricPoller) updateUtilMetrics(group *sync.WaitGroup) {
	defer group.Done()
	p.logger.Debug("start update util metrics")

	v, _ := mem.VirtualMemory()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics.Gauges[metric.TotalMemory] = metric.Gauge(v.Total)
	p.metrics.Gauges[metric.FreeMemory] = metric.Gauge(v.Free)

	p.logger.Debug("successfully updated util metrics")

}
