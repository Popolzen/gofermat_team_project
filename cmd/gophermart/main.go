package main

import (
	"fmt"

	gmconfig "github.com/Popolzen/gofermat_team/internal/gophermart/config"
	gmdb "github.com/Popolzen/gofermat_team/internal/gophermart/db"
)

func main() {
	cfg := gmconfig.NewConfig()
	dbCfg := gmdb.NewDBConfig(*cfg)
	dbCfg.DBurl = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, 5432, `postgres`, `123456`, `shortener`)
	db, err := gmdb.NewDataBase(*cfg, dbCfg)
	if err != nil {
		fmt.Print(err)
	}
	err = db.Migrate()
	if err != nil {
		fmt.Print(err)
	}

	// db.CreateUser(context.Background(), "qwe", "zxcbwd")
	// user, err := db.GetUserByLogin(context.Background(), "qwe")
	// fmt.Print(user, err)
}
