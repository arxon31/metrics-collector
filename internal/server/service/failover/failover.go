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

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/arxon31/metrics-collector/internal/entity"
)

type repo interface {
	// Metrics returns all metrics values
	Metrics(ctx context.Context) ([]entity.MetricDTO, error)
	// StoreBatch stores batch of metrics
	StoreBatch(ctx context.Context, metrics []entity.MetricDTO) error
}

type service struct {
	repo         repo
	path         string
	dumpInterval time.Duration
	isRestore    bool
	logger       *zap.SugaredLogger
}

func NewService(repo repo, path string, dumpInterval time.Duration, isRestore bool, logger *zap.SugaredLogger) *service {
	return &service{
		repo:         repo,
		path:         path,
		dumpInterval: dumpInterval,
		isRestore:    isRestore,
		logger:       logger,
	}
}

func (s *service) Run(ctx context.Context) {
	if s.isRestore {
		if err := s.restore(ctx); err != nil {
			s.logger.Errorln("can not restore data:", err)
		} else {
			s.logger.Infoln("restored data from:", s.path)
		}
	}

	dumper := errgroup.Group{}

	if s.dumpInterval > 0 {
		dumper.Go(func() error {
			return s.dumpByInterval(ctx)
		})
	} else {
		dumper.Go(func() error {
			return s.dumpSynchronously(ctx)
		})
	}

	err := dumper.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		s.logger.Errorln("dumper failed:", err)
	}

	err = s.dump()
	if err != nil {
		s.logger.Errorln("can not dump data after shutdown:", err)
		return
	}
	s.logger.Infoln("dumped data after shutdown to:", s.path)

}

func (s *service) dumpByInterval(ctx context.Context) error {
	ticker := time.NewTicker(s.dumpInterval)
	s.logger.Infof("dumping data by interval %s to %s", s.dumpInterval, s.path)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.dump(); err != nil {
				s.logger.Errorln("can not dump data:", err)
				return err
			}
			s.logger.Infoln("TICK:dumped data to:", s.path)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *service) dumpSynchronously(ctx context.Context) error {
	s.logger.Infof("dumping data synchronously to %s", s.path)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := s.dump()
			if err != nil {
				s.logger.Errorln("can not dump data:", err)
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

	file, err := os.OpenFile(s.path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
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

	err = s.repo.StoreBatch(ctx, metrics)
	if err != nil {
		return err
	}

	return nil
}
