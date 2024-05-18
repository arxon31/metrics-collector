// Package poller polls metrics and returns them to generator for generating http requests
package poller

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/entity"
	"math/rand"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

type repo interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
}

var errMetricSave = "metric save error"

type metricPoller struct {
	mu     sync.Mutex
	repo   repo
	logger *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger, repo repo) *metricPoller {

	p := &metricPoller{
		repo:   repo,
		logger: logger,
	}

	return p

}

// Poll function polls metrics and returns them in Metrics struct
func (p *metricPoller) Poll() {
	p.logger.Debug("start polling metrics")
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go p.updateRuntimeMetrics(wg)
	go p.updateUtilMetrics(wg)

	wg.Wait()
	p.logger.Debug("finished polling metrics")

}

func (p *metricPoller) updateRuntimeMetrics(wg *sync.WaitGroup) {
	ctx := context.Background()
	defer wg.Done()
	p.logger.Debug("start update runtime metrics")

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.repo.Replace(ctx, entity.Alloc, float64(ms.Alloc)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.BuckHashSys, float64(ms.BuckHashSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.Frees, float64(ms.Frees)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.GCCPUFraction, ms.GCCPUFraction); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.GCSys, float64(ms.GCSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.HeapAlloc, float64(ms.HeapAlloc)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.HeapIdle, float64(ms.HeapIdle)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.HeapInuse, float64(ms.HeapInuse)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.HeapObjects, float64(ms.HeapObjects)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.HeapReleased, float64(ms.HeapReleased)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.HeapSys, float64(ms.HeapSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.LastGC, float64(ms.LastGC)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.Lookups, float64(ms.Lookups)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.MCacheInuse, float64(ms.MCacheInuse)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.MCacheSys, float64(ms.MCacheSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.MSpanInuse, float64(ms.MSpanInuse)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.MSpanSys, float64(ms.MSpanSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.Mallocs, float64(ms.Mallocs)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.NextGC, float64(ms.NextGC)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.NumForcedGC, float64(ms.NumForcedGC)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.NumGC, float64(ms.NumGC)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.OtherSys, float64(ms.OtherSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.PauseTotalNs, float64(ms.PauseTotalNs)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.StackInuse, float64(ms.StackInuse)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.StackSys, float64(ms.StackSys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.Sys, float64(ms.Sys)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.TotalAlloc, float64(ms.TotalAlloc)); err != nil {
		p.logger.Error(errMetricSave)
	}

	if err := p.repo.Count(ctx, entity.PollCount, 1); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.RandomValue, rand.Float64()); err != nil {
		p.logger.Error(errMetricSave)
	}

	p.logger.Debug("successfully updated runtime metrics")
}

func (p *metricPoller) updateUtilMetrics(wg *sync.WaitGroup) {
	defer wg.Done()
	p.logger.Debug("start update util metrics")

	v, _ := mem.VirtualMemory()

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.repo.Replace(context.Background(), entity.TotalMemory, float64(v.Total)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(context.Background(), entity.FreeMemory, float64(v.Free)); err != nil {
		p.logger.Error(errMetricSave)
	}

	p.logger.Debug("successfully updated util metrics")

}
