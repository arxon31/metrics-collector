package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arxon31/metrics-collector/pkg/logger"

	"github.com/arxon31/metrics-collector/internal/entity"
	"github.com/arxon31/metrics-collector/internal/repository/repoerr"

	_ "github.com/jackc/pgx/stdlib"
)

type migrationsUpper interface {
	up(db *sql.DB)
}

type Postgres struct {
	db  *sql.DB
	url string
}

const (
	retryAttempts = 3
	startSleep    = 1 * time.Second
)

func NewPostgres(url string) (*Postgres, error) {

	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	migrationsUp(db)

	psql := &Postgres{
		db:  db,
		url: url,
	}

	return psql, nil
}

func (s *Postgres) StoreBatch(ctx context.Context, metrics []entity.MetricDTO) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	gaugesQuery := `INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=$2`
	countersQuery := `INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=counters.value+$2`

	for _, m := range metrics {
		switch m.MetricType {
		case entity.GaugeType:
			_, err = tx.ExecContext(ctx, gaugesQuery, m.Name, *m.Gauge)
			if err != nil {
				return err
			}
		case entity.CounterType:
			_, err = tx.ExecContext(ctx, countersQuery, m.Name, *m.Counter)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Postgres) StoreGauge(ctx context.Context, name string, value float64) error {
	query := `INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=$2`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, value)
	if err != nil {
		err = s.retryStore(retryAttempts, stmt, name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Postgres) StoreCounter(ctx context.Context, name string, value int64) error {
	query := `INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=counters.value+$2`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, value)
	if err != nil {
		err = s.retryStore(retryAttempts, stmt, name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Postgres) Gauge(ctx context.Context, name string) (float64, error) {
	query := `SELECT value FROM gauges WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val float64
	err := row.Scan(&val)
	if err != nil {
		return 0, repoerr.ErrMetricNotFound
	}
	return val, nil
}
func (s *Postgres) Counter(ctx context.Context, name string) (int64, error) {
	query := `SELECT value FROM counters WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val int64
	err := row.Scan(&val)
	if err != nil {
		return 0, repoerr.ErrMetricNotFound
	}
	return val, nil
}
func (s *Postgres) Metrics(ctx context.Context) ([]entity.MetricDTO, error) {
	query := `SELECT (name, value) FROM gauges;`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	metrics := make([]entity.MetricDTO, 0)

	for rows.Next() {
		var gaugeMetric = entity.MetricDTO{MetricType: entity.GaugeType}

		err = rows.Scan(&gaugeMetric.Name, gaugeMetric.Gauge)
		if err != nil {
			logger.Logger.Error(err)
			continue
		}

		metrics = append(metrics, gaugeMetric)
	}

	err = rows.Err()
	if err != nil {
		logger.Logger.Error(err)
	}

	query = `SELECT (name, value) FROM counters;`
	rows, err = s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var counterMetric = entity.MetricDTO{MetricType: entity.CounterType}
		err = rows.Scan(&counterMetric.Name, &counterMetric.Counter)
		if err != nil {
			logger.Logger.Error(err)
			continue
		}

		metrics = append(metrics, counterMetric)
	}

	err = rows.Err()
	if err != nil {
		logger.Logger.Error(err)
	}

	return metrics, nil
}

func (s *Postgres) Ping() error {
	err := s.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (s *Postgres) retryStore(attempts int, stmt *sql.Stmt, name string, value any) (err error) {
	sleep := startSleep
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep += 2 * time.Second
		}
		_, err = stmt.Exec(name, value)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, error: %s", attempts, err)
}
