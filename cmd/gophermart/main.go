package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	gmconfig "github.com/Popolzen/gofermat_team/internal/gophermart/config"
	gmdb "github.com/Popolzen/gofermat_team/internal/gophermart/db"
	gmhandlers "github.com/Popolzen/gofermat_team/internal/gophermart/handlers/auth"
	gmauth "github.com/Popolzen/gofermat_team/internal/gophermart/middleware/auth"
	gmservice "github.com/Popolzen/gofermat_team/internal/gophermart/service"
	gmstorage "github.com/Popolzen/gofermat_team/internal/gophermart/storage/database"
)

func main() {
	// Загрузка конфига (расширь под ТЗ: RUN_ADDRESS, etc.)
	cfg := gmconfig.NewConfig()
	jwtSecret := "your-jwt-secret" // В prod из env

	// DB setup
	dbCfg := gmdb.NewDBConfig(*cfg)
	dbCfg.DBurl = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, 5432, `postgres`, `123456`, `gophermart`) // Имя БД под ТЗ
	db, err := gmdb.NewDataBase(*cfg, dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Migrate(); err != nil {
		log.Fatal(err)
	}

	// Storage (один для всех, реализует мелкие интерфейсы)
	storage := gmstorage.NewPostgresStorage(db.DB)

	// Auth
	auth := gmauth.NewAuth(jwtSecret)

	// Services (UserService для register)
	userService := gmservice.NewUserService(storage, auth)

	// Router (chi для удобства)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Route("/user", func(user chi.Router) {
			user.Post("/register", gmhandlers.RegisterHandler(userService))
			user.Post("/login", gmhandlers.LoginHandler(userService)) // Если есть, добавь
		})
	})

	// Health check или root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Gophermart API running"))
	})

	// Запуск сервера
	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	log.Printf("Starting server on %s", cfg.ServerAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
