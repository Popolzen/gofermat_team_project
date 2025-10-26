package repository

import (
	"context"

	"github.com/Popolzen/gofermat_team/internal/accrual/model"
)

type AccrualRepository interface {
	GetOrderAccrual(ctx context.Context, order string) (*model.OrderAccrual, error)
	OrderExists(ctx context.Context, orderNumber string) (bool, error)
	CreateOrder(ctx context.Context, orderNumber string, status model.Status) error
	UpdateOrderAccrual(ctx context.Context, orderNumber string, accrual float64, status model.Status) error
	UpdateOrderStatus(ctx context.Context, orderNumber string, status model.Status) error
	FindRewardByDescription(ctx context.Context, description string) (*model.GoodReward, error)
	CreateGoodsReward(ctx context.Context, reward *model.GoodReward) error
}
