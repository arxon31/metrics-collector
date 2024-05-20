// Package poller polls metrics and returns them to generator for generating http requests
package poller

import (
	"context"
	"github.com/arxon31/metrics-collector/internal/agent/config"
	"github.com/arxon31/metrics-collector/internal/agent/service/generator"
	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

type pollerRepo interface {
	Replace(ctx context.Context, name string, value float64) error
	Count(ctx context.Context, name string, value int64) error
}

type generatorRepo interface {
	StoreCounter(ctx context.Context, name string, value int64) error
	Metrics(ctx context.Context) ([]entity.MetricDTO, error)
}

type hasher interface {
	Hash([]byte) (string, error)
}

type compressor interface {
	Compress([]byte) ([]byte, error)
}

type requestGenerator interface {
	Generate(ctx context.Context)
	Requests() chan *resty.Request
}

var errMetricSave = "metric save error"

type metricPoller struct {
	repo         pollerRepo
	gen          requestGenerator
	pollInterval time.Duration
	logger       *zap.SugaredLogger
}

func New(logger *zap.SugaredLogger, pRepo pollerRepo, gRepo generatorRepo, cfg *config.Config, hasher hasher, compressor compressor) chan *resty.Request {

	p := &metricPoller{
		gen:    generator.New(cfg.Address, cfg.RateLimit, gRepo, hasher, compressor, logger),
		repo:   pRepo,
		logger: logger,
	}

	return p.gen.Requests()

}

func (p *metricPoller) Run(ctx context.Context) {

	pollTimer := time.NewTicker(p.pollInterval)
	defer pollTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("poller gracefully stopped")
			return
		case <-pollTimer.C:
			p.poll(ctx)
			p.gen.Generate(ctx)
		}
	}
}

// poll function polls metrics and returns them in Metrics struct
func (p *metricPoller) poll(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go p.updateRuntimeMetrics(ctx, wg)
	go p.updateUtilMetrics(ctx, wg)

	wg.Wait()

}

func (p *metricPoller) updateRuntimeMetrics(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	p.logger.Debug("start update runtime metrics")

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

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

func (p *metricPoller) updateUtilMetrics(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	p.logger.Debug("start update util metrics")

	v, _ := mem.VirtualMemory()

	if err := p.repo.Replace(ctx, entity.TotalMemory, float64(v.Total)); err != nil {
		p.logger.Error(errMetricSave)
	}
	if err := p.repo.Replace(ctx, entity.FreeMemory, float64(v.Free)); err != nil {
		p.logger.Error(errMetricSave)
	}

	p.logger.Debug("successfully updated util metrics")

}
