package handler

import (
	"github.com/go-chi/chi/v5"
	myLogger "github.com/Popolzen/gofermat_team/internal/accrual/logger"
)

func (h *accrualHandler) InitRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(myLogger.NewLoggerMiddleware(h.logger))

	r.Route("/api", func(r chi.Router) {
		r.With(h.rateLimitMiddleware()).Get("/orders/{number}", h.GetOrder)
		r.With(h.rateLimitMiddleware()).Post("/orders", h.RegisterOrder)
		r.With(h.rateLimitMiddleware()).Post("/goods", h.RegisterGoodsReward)
	})

	return r
}
