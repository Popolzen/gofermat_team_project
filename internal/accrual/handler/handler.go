package handler

import (
	"sync"

	"github.com/Popolzen/gofermat_team/internal/accrual/config"
	"github.com/Popolzen/gofermat_team/internal/accrual/db"
	"github.com/Popolzen/gofermat_team/internal/accrual/logger"
	"github.com/Popolzen/gofermat_team/internal/accrual/service"
	"golang.org/x/time/rate"
)

type accrualHandler struct {
	accrualUseCase service.AccrualUseCase 
	rateLimiters   map[string]*rate.Limiter
	mutex          sync.Mutex
	config         *config.ServiceConfig
	logger         *logger.Logger
	dbCfg          *db.DBConfig
}

func NewAccrualHandler(
	accrualUseCase service.AccrualUseCase,  
	cfg *config.ServiceConfig, 
	logger *logger.Logger, 
	dbCfg *db.DBConfig,
) *accrualHandler {
	return &accrualHandler{
		accrualUseCase: accrualUseCase,  
		rateLimiters:   make(map[string]*rate.Limiter),
		config:         cfg,
		logger:         logger,
		dbCfg:          dbCfg,
	}
}

