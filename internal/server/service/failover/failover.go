package failover

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/entity"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"path/filepath"
	"time"
)

type repo interface {
	// Metrics returns all metrics values
	Metrics(ctx context.Context) (string, error)
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
		}
		s.logger.Infoln("restored data from:", s.path)
	}

	dumper := errgroup.Group{}

	if s.dumpInterval > 0 {
		dumper.Go(func() error {
			return s.dumpByInterval(ctx)
		})
	} else {
		dumper.Go(func() error {
			return s.dumpByInterval(ctx)
		})
	}

	err := dumper.Wait()
	if err != nil {
		s.logger.Errorln("dumper failed:", err)
	}
}

func (s *service) dumpByInterval(ctx context.Context) error {
	ticker := time.NewTicker(s.dumpInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.dump(); err != nil {
				s.logger.Errorln("can not dump data:", err)
				return err
			}
		case <-ctx.Done():
			err := s.dump()
			if err != nil {
				s.logger.Errorln("can not dump data after app shutdown:", err)
				return err
			}
			s.logger.Infoln("dumped data after app shutdown to:", s.path)
		}
	}
}

func (s *service) dumpSynchronously(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			err := s.dump()
			if err != nil {
				s.logger.Errorln("can not dump data after app shutdown:", err)
				return err
			}
			s.logger.Infoln("dumped data after app shutdown to:", s.path)
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

	_, err = file.Write([]byte(metrics))
	if err != nil {
		return err
	}

	return nil
}

func (s *service) restore(ctx context.Context) error {
	file, err := os.OpenFile(s.path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can not open restoring file: %w", err)
	}
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
