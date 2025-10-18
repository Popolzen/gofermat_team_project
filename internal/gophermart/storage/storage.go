package gmstorage

import (
	"context"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
)

// UserStorage - методы для работы с пользователями
type UserStorage interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (*gmmodel.User, error)
}

// OrderStorage - методы для работы с заказами
type OrderStorage interface {
	CreateOrder(ctx context.Context, userID int64, orderNumber string) error
	GetOrderByNumber(ctx context.Context, orderNumber string) (*gmmodel.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]*gmmodel.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNumber, status string, accrual *float64) error
	GetOrdersForProcessing(ctx context.Context) ([]*gmmodel.Order, error)
}

// BalanceStorage - методы для работы с балансом
type BalanceStorage interface {
	GetUserBalance(ctx context.Context, userID int64) (*gmmodel.Balance, error)
}

// WithdrawalStorage - методы для работы со списаниями
type WithdrawalStorage interface {
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]*gmmodel.Withdrawal, error)
}

// Storage - полный интерфейс, объединяющий все операции
type Storage interface {
	UserStorage
	OrderStorage
	BalanceStorage
	WithdrawalStorage
}
