package postgres

import (
	"context"
	"database/sql"
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

	queryGauges := `CREATE TABLE IF NOT EXISTS gauges(name text primary key, value double precision)`
	queryCounters := `CREATE TABLE IF NOT EXISTS counters(name text primary key, value int)`

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

func (s *PSQL) Replace(ctx context.Context, name string, value float64) error {
	query := `UPDATE gauges SET value=$1 WHERE name=$2;`
	_, err := s.db.ExecContext(ctx, query, value, name)
	if err != nil {
		return err
	}
	return nil

}
func (s *PSQL) Count(ctx context.Context, name string, value int64) error {
	query := `SELECT value FROM counters WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val int64
	err := row.Scan(&val)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			query = `INSERT INTO counters (name, value) VALUES ($1, $2)`
			_, err = s.db.ExecContext(ctx, query, name, value)
			if err != nil {
				return err
			}
			return nil
		default:
			return err
		}
	}

	value += val

	query = `UPDATE counters SET value=$1 WHERE name=$2`
	_, err = s.db.ExecContext(ctx, query, value, name)
	if err != nil {
		return err
	}

	return nil
}
func (s *PSQL) GaugeValue(ctx context.Context, name string) (float64, error) {
	query := `SELECT value FROM gauges WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val float64
	err := row.Scan(&val)
	if err != nil {
		return 0, err
	}
	return val, nil
}
func (s *PSQL) CounterValue(ctx context.Context, name string) (int64, error) {
	query := `SELECT value FROM counters WHERE name=$1;`
	row := s.db.QueryRowContext(ctx, query, name)

	var val int64
	err := row.Scan(&val)
	if err != nil {
		return 0, err
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
