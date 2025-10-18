package gmdb

import (
	"database/sql"
	"fmt"

	gmconfig "github.com/Popolzen/gofermat_team/internal/gophermart/config"
	gmmigration "github.com/Popolzen/gofermat_team/migrations/gophermart"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBConfig содержит конфигурацию для подключения к БД
type DBConfig struct {
	DBurl string
}

// Database представляет подключение к базе данных
type Database struct {
	*sql.DB
	config *DBConfig
}

// NewDBConfig создает новую конфигурацию БД
func NewDBConfig(c gmconfig.Config) DBConfig {
	return DBConfig{
		DBurl: c.DBurl,
	}
}

// NewDataBase создает абстракцию БД
func NewDataBase(c gmconfig.Config, dbConf DBConfig) (*Database, error) {
	db, err := sql.Open("pgx", dbConf.DBurl)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть подключение: %w", err)
	}
	return &Database{
		DB:     db,
		config: &dbConf,
	}, nil
}

func (d *Database) Migrate() error {
	return gmmigration.MigrateUp(d.DB)
}

// Close закрывает подключение к БД
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}
