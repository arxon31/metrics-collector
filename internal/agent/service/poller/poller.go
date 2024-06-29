// Package poller polls metrics and returns them to generator for generating http requests
package poller

import (
	"context"
	"github.com/arxon31/metrics-collector/pkg/logger"
	"math/rand"
	"runtime"
	"time"

	"github.com/arxon31/metrics-collector/internal/entity"

	"github.com/shirou/gopsutil/mem"
)

type pollerRepo interface {
	StoreGauge(ctx context.Context, name string, value float64) error
	StoreCounter(ctx context.Context, name string, value int64) error
}

var errMetricSave = "metric save error"

type metricPoller struct {
	repo         pollerRepo
	pollInterval time.Duration
}

// New creates new poller
func New(pRepo pollerRepo) *metricPoller {

	p := &metricPoller{
		repo: pRepo,
	}

	return p
}

// Poll function polls metrics and returns them in Metrics struct
func (p *metricPoller) Poll(ctx context.Context) {
	p.updateRuntimeMetrics(ctx)
	p.updateUtilMetrics(ctx)
}

func (p *metricPoller) updateRuntimeMetrics(ctx context.Context) {
	logger.Logger.Debug("start update runtime metrics")

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	if err := p.repo.StoreGauge(ctx, entity.Alloc, float64(ms.Alloc)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.BuckHashSys, float64(ms.BuckHashSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.Frees, float64(ms.Frees)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.GCCPUFraction, ms.GCCPUFraction); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.GCSys, float64(ms.GCSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.HeapAlloc, float64(ms.HeapAlloc)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.HeapIdle, float64(ms.HeapIdle)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.HeapInuse, float64(ms.HeapInuse)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.HeapObjects, float64(ms.HeapObjects)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.HeapReleased, float64(ms.HeapReleased)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.HeapSys, float64(ms.HeapSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.LastGC, float64(ms.LastGC)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.Lookups, float64(ms.Lookups)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.MCacheInuse, float64(ms.MCacheInuse)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.MCacheSys, float64(ms.MCacheSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.MSpanInuse, float64(ms.MSpanInuse)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.MSpanSys, float64(ms.MSpanSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.Mallocs, float64(ms.Mallocs)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.NextGC, float64(ms.NextGC)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.NumForcedGC, float64(ms.NumForcedGC)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.NumGC, float64(ms.NumGC)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.OtherSys, float64(ms.OtherSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.PauseTotalNs, float64(ms.PauseTotalNs)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.StackInuse, float64(ms.StackInuse)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.StackSys, float64(ms.StackSys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.Sys, float64(ms.Sys)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.TotalAlloc, float64(ms.TotalAlloc)); err != nil {
		logger.Logger.Error(errMetricSave)
	}

	if err := p.repo.StoreCounter(ctx, entity.PollCount, 1); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.RandomValue, rand.Float64()); err != nil {
		logger.Logger.Error(errMetricSave)
	}

	logger.Logger.Debug("successfully updated runtime metrics")
}

func (p *metricPoller) updateUtilMetrics(ctx context.Context) {
	logger.Logger.Debug("start update util metrics")

	v, _ := mem.VirtualMemory()

	if err := p.repo.StoreGauge(ctx, entity.TotalMemory, float64(v.Total)); err != nil {
		logger.Logger.Error(errMetricSave)
	}
	if err := p.repo.StoreGauge(ctx, entity.FreeMemory, float64(v.Free)); err != nil {
		logger.Logger.Error(errMetricSave)
	}

	logger.Logger.Debug("successfully updated util metrics")

}
