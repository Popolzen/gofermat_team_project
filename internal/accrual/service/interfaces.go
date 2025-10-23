package service

import (
	"context"

	"github.com/Popolzen/gofermat_team/internal/accrual/model"
)

type AccrualUseCase interface {
	GetOrderAccrual(ctx context.Context, order string) (*model.OrderAccrual, error)
	RegisterOrder(ctx context.Context, order *model.OrderRequest) error
	Stop()
	CalculateOrderAccrual(ctx context.Context, goods []model.Good) (float64, error)
	RegisterGoodsReward(ctx context.Context, reward *model.GoodReward) error
}
