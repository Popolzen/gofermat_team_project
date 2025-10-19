package gmmodel

import (
	"errors"
	"time"
)

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}

type Order struct {
	ID          int64
	UserID      int64
	OrderNumber string
	Status      string
	Accrual     *float64
	UploadedAt  time.Time
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

type Withdrawal struct {
	OrderNumber string
	Sum         float64
	ProcessedAt time.Time
}

var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginAlreadyExists = errors.New("login already exists")
	ErrInvalidPassword    = errors.New("invalid password")

	// Order errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrOrderOwnedByOther  = errors.New("order owned by another user")
	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrNoOrders           = errors.New("no orders found")

	// Balance errors
	ErrInsufficientFunds = errors.New("insufficient funds")
)

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)
