package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Загрузка конфига
	cfg := gmconfig.NewConfig()
	jwtSecret := "your-jwt-secret" // В prod из env

	// DB setup
	dbCfg := gmdb.NewDBConfig(*cfg)
	dbCfg.DBurl = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		`localhost`, 5432, `postgres`, `123456`, `shortener`)
	db, err := gmdb.NewDataBase(*cfg, dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Migrate(); err != nil {
		log.Fatal(err)
	}

	// Storage
	storage := gmstorage.NewPostgresStorage(db.DB)

	// Auth
	auth := gmauth.NewAuth(jwtSecret)

	// Services
	userService := gmservice.NewUserService(storage, auth)
	userOrder := gmservice.NewOrderService(storage)

	// OrderProcessor для опроса и обновления заказов
	accrualURL := "http://localhost:8081" // поменяй меня!!! cfg.AccrualURL
	processor := gmservice.NewOrderProcessor(storage, accrualURL)

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск процессора в фоне
	go processor.Start(ctx)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Route("/user", func(user chi.Router) {
			// Публичные роуты (без auth)
			user.Post("/register", gmhandlers.RegisterHandler(userService))
			user.Post("/login", gmhandlers.LoginHandler(userService))

			// Защищённые роуты (с auth middleware)
			user.Group(func(protected chi.Router) {
				protected.Use(auth.AuthMiddleware) // Применяем auth middleware ко всем роутам в группе
				// Заказы
				protected.Post("/orders", gmhandlers.UploadHandler(userOrder))
				protected.Get("/orders", gmhandlers.GetOrdersHandler(userOrder))
			})
		})
	})

	// Health check или root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Gophermart API running"))
	})

	// Запуск сервера
	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Сервер в фоне
	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Остановка процессора
	processor.Stop()
	cancel()

	// Shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
