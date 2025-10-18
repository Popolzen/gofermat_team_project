package gmstorage

import gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"

type Storage interface {
	// Users
	CreateUser(login, passwordHash string) (int64, error)
	GetUserByLogin(login string) (*gmmodel.User, error)

	// Orders
	CreateOrder(userID int64, orderNumber string) error
	GetOrderByNumber(orderNumber string) (*gmmodel.Order, error)
	GetUserOrders(userID int64) ([]*gmmodel.Order, error)
	UpdateOrderStatus(orderNumber, status string, accrual *float64) error
	GetOrdersForProcessing() ([]*gmmodel.Order, error)

	// Balance
	GetUserBalance(userID int64) (*gmmodel.Balance, error)

	// Withdrawals
	CreateWithdrawal(userID int64, orderNumber string, sum float64) error
	GetUserWithdrawals(userID int64) ([]*gmmodel.Withdrawal, error)
}
