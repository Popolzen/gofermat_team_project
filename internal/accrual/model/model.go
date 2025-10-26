package model

import "errors"

type Status string

const (
	Registered Status = "REGISTERED"
	Processing Status = "PROCESSING"
	Invalid    Status = "INVALID"
	Processed  Status = "PROCESSED"
)

type OrderAccrual struct {
	Order   string   `json:"order"`
	Status  Status   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type OrderRequest struct {
	Order string `json:"order"`
	Goods []Good `json:"goods"`
}

type Good struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type GoodReward struct {
	Description string  `json:"description"`
	RewardType  string  `json:"reward_type"`
	RewardValue float64 `json:"reward_value"`
	Match       string  `json:"match"`
}

type ProcessingResult struct {
	OrderNumber string
	Accrual     float64
	Status      Status
	Error       error
}

var (
	ErrNotFound            = errors.New("order not registered")
	ErrInvalidOrder        = errors.New("invalid order number")
	ErrInvalidOrderFormat  = errors.New("invalid order format")
	ErrOrderAlreadyExists  = errors.New("order already exists")
	ErrRewardAlreadyExists = errors.New("reward already exists")
	ErrInvalidRewardFormat = errors.New("invalid reward format")
)
