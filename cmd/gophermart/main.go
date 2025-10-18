package main

import (
	gmconfig "github.com/Popolzen/gofermat_team/internal/gophermart/config"
	gmdb "github.com/Popolzen/gofermat_team/internal/gophermart/db"
)

func main() {
	cfg := gmconfig.NewConfig()
	dbCfg := gmdb.NewDBConfig(*cfg)
	db, err := gmdb.NewDataBase(*cfg, dbCfg)
	if err != nil {

	}
	db.Migrate()
}
