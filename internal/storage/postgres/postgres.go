package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/arxon31/metrics-collector/internal/storage/mem"
	"github.com/arxon31/metrics-collector/pkg/metric"
	_ "github.com/jackc/pgx/stdlib"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type PSQL struct {
	db     *sql.DB
	conn   string
	logger *zap.SugaredLogger
}

type RetryableError struct {
	err        error
	Timer      time.Time
	RetryCount int
}

func NewPostgres(conn string, logger *zap.SugaredLogger) (*PSQL, error) {

	db, err := sql.Open("pgx", conn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queryGauges := `CREATE TABLE IF NOT EXISTS gauges (
    		name text PRIMARY KEY NOT NULL, 
    		value DOUBLE PRECISION
    )`
	queryCounters := `CREATE TABLE IF NOT EXISTS counters (
    		name text PRIMARY KEY NOT NULL, 
    		value BIGINT
    )`

	_, err = db.ExecContext(ctx, queryGauges)
	if err != nil {
		logger.Errorln("can not create table for gauge metrics due to error:", err)
		return &PSQL{}, err
	}

	_, err = db.ExecContext(ctx, queryCounters)
	if err != nil {
		logger.Errorln("can not create table for counter metrics due to error:", err)
		return &PSQL{}, err
	}

	psql := &PSQL{
		db:     db,
		conn:   conn,
		logger: logger,
	}

	return psql, nil
}

func (s *PSQL) StoreBatch(ctx context.Context, metrics []metric.MetricDTO) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	gaugesQuery := `INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=$2`
	countersQuery := `INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=counters.value+$2`

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			_, err = tx.ExecContext(ctx, gaugesQuery, m.ID, *m.Value)
			if err != nil {
				return err
			}
		case "counter":
			_, err = tx.ExecContext(ctx, countersQuery, m.ID, *m.Delta)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *PSQL) Replace(ctx context.Context, name string, value float64) error {
	err := s.replace(ctx, name, value)
	if err != nil {
		err = s.retryReplace(3, s.replace, ctx, name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PSQL) replace(ctx context.Context, name string, value float64) error {
	query := `INSERT INTO gauges (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=$2`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *PSQL) retryReplace(attempts int, f func(ctx context.Context, name string, value float64) error, ctx context.Context, name string, value float64) (err error) {
	sleep := 1 * time.Second
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep += 2 * time.Second
		}
		err = f(ctx, name, value)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, error: %s", attempts, err)
}

func (s *PSQL) Count(ctx context.Context, name string, value int64) error {
	err := s.count(ctx, name, value)
	if err != nil {
		err = s.retryCount(3, s.count, ctx, name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PSQL) count(ctx context.Context, name string, value int64) error {
	query := `INSERT INTO counters (name, value) VALUES ($1, $2) ON CONFLICT (name) DO UPDATE SET value=counters.value+$2`
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, value)
	if err != nil {
		return err
	}

	return nil
}

func (s *PSQL) retryCount(attempts int, f func(ctx context.Context, name string, value int64) error, ctx context.Context, name string, value int64) (err error) {
	sleep := 1 * time.Second
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(sleep)
			sleep += 2 * time.Second
		}
		err = f(ctx, name, value)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, error: %s", attempts, err)
}

func (s *PSQL) GaugeValue(ctx context.Context, name string) (float64, error) {
	query := `SELECT value FROM gauges WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val float64
	err := row.Scan(&val)
	if err != nil {
		return 0, mem.ErrIsNotFound
	}
	return val, nil
}
func (s *PSQL) CounterValue(ctx context.Context, name string) (int64, error) {
	query := `SELECT value FROM counters WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val int64
	err := row.Scan(&val)
	if err != nil {
		return 0, mem.ErrIsNotFound
	}
	return val, nil
}
func (s *PSQL) Values(ctx context.Context) (string, error) {
	query := `SELECT * FROM gauges;`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}

	var res strings.Builder

	for rows.Next() {
		var name string
		var val float64

		err = rows.Scan(&name, &val)
		if err != nil {
			return "", err
		}

		valString := strconv.FormatFloat(val, 'g', -1, 64)

		nameValString := strings.Join([]string{name, valString}, " ")

		res.WriteString(nameValString)
	}

	err = rows.Err()
	if err != nil {
		return "", err
	}

	query = `SELECT * FROM counters;`
	rows, err = s.db.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}

	for rows.Next() {
		var name string
		var val int64

		err = rows.Scan(&name, &val)
		if err != nil {
			return "", err
		}

		valString := strconv.FormatInt(val, 10)

		nameValString := strings.Join([]string{name, valString}, " ")

		res.WriteString(nameValString)
	}

	err = rows.Err()
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func (s *PSQL) Ping() error {
	db, err := sql.Open("pgx", s.conn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (s *PSQL) Dump(ctx context.Context, path string) error {
	return nil
}

func (s *PSQL) Restore(ctx context.Context, path string) error {
	return nil
}
