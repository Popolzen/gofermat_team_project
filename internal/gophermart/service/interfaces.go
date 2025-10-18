package service

import (
	"context"
	"time"
)

type UserService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
}

type OrderService interface {
	Upload(ctx context.Context, userID int64, orderNumber string) error
	GetList(ctx context.Context, userID int64) ([]*OrderDTO, error)
}
type OrderDTO struct {
	Number     string
	Status     string
	Accrual    *float64
	UploadedAt time.Time
}

type BalanceService interface {
	Get(ctx context.Context, userID int64) (*BalanceDTO, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*WithdrawalDTO, error)
}
type BalanceDTO struct {
	Current   float64
	Withdrawn float64
}

type WithdrawalDTO struct {
	Order       string
	Sum         float64
	ProcessedAt time.Time
}
