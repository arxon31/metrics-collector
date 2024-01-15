package postgres

import (
	"database/sql"
	_ "github.com/jackc/pgx/stdlib"
)

type PSQL struct {
	db   *sql.DB
	conn string
}

func NewPostgres(conn string) *PSQL {
	return &PSQL{
		conn: conn,
	}
}

func (s *PSQL) Ping() error {
	db, err := sql.Open("pgx", s.conn)
	if err != nil {
		return err
	}
	s.db = db

	err = s.db.Ping()
	if err != nil {
		return err
	}
	return nil
}
