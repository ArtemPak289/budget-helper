package repository

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

func OpenSQLite(path string) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)
	if path == "" {
		err = errors.New("database path is required")
	}
	if err == nil {
		db, err = sql.Open("sqlite", path)
	}
	if err == nil {
		err = db.Ping()
	}
	return db, err
}
