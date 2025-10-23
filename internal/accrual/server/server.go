package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Popolzen/gofermat_team/internal/accrual/config"
	"github.com/Popolzen/gofermat_team/internal/accrual/db"
	"github.com/Popolzen/gofermat_team/internal/accrual/handler"
	"github.com/Popolzen/gofermat_team/internal/accrual/logger"
	"github.com/Popolzen/gofermat_team/internal/accrual/repository"
	"github.com/Popolzen/gofermat_team/internal/accrual/service"
	migrations "github.com/Popolzen/gofermat_team/migrations/accrual"
	"go.uber.org/zap"
)

func Execute() {
	cfg := config.NewConfig()

	myLogger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer myLogger.Sync()

	myLogger.Info("Starting accrual service",
		zap.String("address", cfg.ServerPort),
		zap.String("database_dsn", cfg.DatabaseDSN),
		zap.String("log_level", cfg.LogLevel.String()),
	)

	dbCfg := db.NewDBConfig(cfg.DatabaseDSN)
	dbInstance, err := db.OpenDB(dbCfg.DSN)
	if err != nil {
		myLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbInstance.Close()

	myLogger.Info("Running database migrations...")
	if err := migrations.MigrateUp(dbInstance); err != nil {
		myLogger.Fatal("Failed to run migrations", zap.Error(err))
	}
	myLogger.Info("Database migrations completed successfully")

	repo := repository.NewPostgresAccrualRepository(dbInstance)
	useCase := service.NewAccrualUseCase(repo, myLogger)

	handler := handler.NewAccrualHandler(useCase, cfg, myLogger, dbCfg)

	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: handler.InitRouter(),
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		myLogger.Info("Server is shutting down...")
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			myLogger.Error("Could not gracefully shutdown the server", zap.Error(err))
		}
		
		useCase.Stop()
		
		close(done)
	}()

	myLogger.Info("Server started successfully", zap.String("address", cfg.ServerPort))
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		myLogger.Fatal("Server failed to start", zap.Error(err))
	}

	<-done
	myLogger.Info("Server stopped")
}
