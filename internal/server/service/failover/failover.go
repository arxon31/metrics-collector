package failover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/arxon31/metrics-collector/pkg/logger"

	"github.com/arxon31/metrics-collector/internal/entity"
)

type repo interface {
	// Metrics returns all metrics values
	Metrics(ctx context.Context) ([]entity.MetricDTO, error)
	StoreGauge(ctx context.Context, name string, value float64) error
	StoreCounter(ctx context.Context, name string, value int64) error
}

type service struct {
	repo         repo
	path         string
	dumpInterval time.Duration
	isRestore    bool
}

func NewService(repo repo, path string, dumpInterval time.Duration, isRestore bool) *service {
	return &service{
		repo:         repo,
		path:         path,
		dumpInterval: dumpInterval,
		isRestore:    isRestore,
	}
}

// Run runs the service with the given context.
func (s *service) Run(ctx context.Context) {
	if s.isRestore {
		if err := s.restore(ctx); err != nil {
			logger.Logger.Errorln("can not restore data:", err)
		} else {
			logger.Logger.Infoln("restored data from:", s.path)
		}
	}

	if s.path == "" {
		return
	}

	if s.dumpInterval > 0 {
		err := s.dumpByInterval(ctx)
		if err != nil {
			logger.Logger.Errorln("can not dump data by interval:", err)
			return
		}

	} else {
		err := s.dumpSynchronously(ctx)
		if err != nil {
			logger.Logger.Errorln("can not dump data synchronously:", err)
			return
		}
	}

}

func (s *service) dumpByInterval(ctx context.Context) error {
	ticker := time.NewTicker(s.dumpInterval)
	logger.Logger.Infof("dumping data by interval %s to %s", s.dumpInterval, s.path)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.dump(); err != nil {
				logger.Logger.Errorln("can not dump data:", err)
				return err
			}
			logger.Logger.Infoln("TICK:dumped data to:", s.path)
		case <-ctx.Done():
			err := s.dump()
			if err != nil {
				logger.Logger.Errorln("can not dump data after shutdown:", err)
				return err
			}
			logger.Logger.Infoln("dumped data after shutdown to:", s.path)
			return nil
		}
	}
}

func (s *service) dumpSynchronously(ctx context.Context) error {
	logger.Logger.Infof("dumping data synchronously to %s", s.path)
	for {
		select {
		case <-ctx.Done():
			err := s.dump()
			if err != nil {
				logger.Logger.Errorln("can not dump data after shutdown:", err)
				return err
			}
			logger.Logger.Infoln("dumped data after shutdown to:", s.path)
			return nil
		default:
			err := s.dump()
			if err != nil {
				logger.Logger.Errorln("can not dump data:", err)
				return err
			}
		}
	}
}

func (s *service) dump() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("can not mkdir: %w", err)
	}

	file, err := os.OpenFile(s.path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("can not open dump file: %w", err)
	}
	defer file.Close()

	metrics, err := s.repo.Metrics(context.Background())
	if err != nil {
		return err
	}

	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("can not marshal to JSON: %w", err)
	}

	_, err = file.Write(metricsJSON)
	if err != nil {
		return fmt.Errorf("can not write to file: %w", err)
	}

	return nil
}

func (s *service) restore(ctx context.Context) error {
	if _, err := os.Stat(s.path); errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("can not open restoring file: %w", err)
	}
	defer file.Close()

	rawMetrics, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("can not read restoring file: %w", err)
	}

	metrics := make([]entity.MetricDTO, 0)

	err = json.Unmarshal(rawMetrics, &metrics)
	if err != nil {
		return fmt.Errorf("can not unmarshal to DTO: %w", err)
	}

	for _, m := range metrics {
		switch m.MetricType {
		case entity.GaugeType:
			err = s.repo.StoreGauge(ctx, m.Name, *m.Gauge)
			if err != nil {
				return fmt.Errorf("can not store gauge: %w", err)
			}
		case entity.CounterType:
			err = s.repo.StoreCounter(ctx, m.Name, *m.Counter)
			if err != nil {
				return fmt.Errorf("can not store counter: %w", err)
			}
		}
	}

	return nil
}
