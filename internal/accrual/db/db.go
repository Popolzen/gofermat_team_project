package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	DSN string
}

func NewDBConfig(dsn string) *DBConfig {
	return &DBConfig{
		DSN: dsn,
	}
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
