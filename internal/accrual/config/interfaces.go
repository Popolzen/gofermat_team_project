package config

import (
    "database/sql"
    "github.com/Popolzen/gofermat_team/internal/accrual/logger"
    "github.com/Popolzen/gofermat_team/internal/accrual/repository"
    "github.com/Popolzen/gofermat_team/internal/accrual/service"
)

type Dependencies struct {
    DB      *sql.DB
    Logger  *logger.Logger
    UseCase service.AccrualUseCase
}

func NewDependencies(db *sql.DB, cfg *ServiceConfig) (*Dependencies, error) {
    logger, err := logger.NewLogger(cfg.LogLevel)
    if err != nil {
        db.Close()
        return nil, err
    }

    repo := repository.NewPostgresAccrualRepository(db)
    useCase := service.NewAccrualUseCase(repo, logger)

    return &Dependencies{
        DB:      db,
        Logger:  logger,
        UseCase: useCase,
    }, nil
}
