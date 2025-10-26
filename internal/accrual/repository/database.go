package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Popolzen/gofermat_team/internal/accrual/model"
)

type PostgresAccrualRepository struct {
	db *sql.DB
}

func NewPostgresAccrualRepository(db *sql.DB) *PostgresAccrualRepository {
	return &PostgresAccrualRepository{db: db}
}

func (r *PostgresAccrualRepository) GetOrderAccrual(ctx context.Context, order string) (*model.OrderAccrual, error) {
	query := `
        SELECT order_number, status, accrual
        FROM accrual_orders
        WHERE order_number = $1
    `
	var accrual model.OrderAccrual
	var dbStatus string
	var accrualVal sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, order).Scan(&accrual.Order, &dbStatus, &accrualVal)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	switch dbStatus {
	case "REGISTERED":
		accrual.Status = model.Registered
	case "PROCESSING":
		accrual.Status = model.Processing
	case "INVALID":
		accrual.Status = model.Invalid
	case "PROCESSED":
		accrual.Status = model.Processed
	default:
		return nil, errors.New("invalid status in database")
	}

	if accrualVal.Valid {
		accrual.Accrual = &accrualVal.Float64
	}
	return &accrual, nil
}

func (r *PostgresAccrualRepository) OrderExists(ctx context.Context, orderNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM accrual_orders WHERE order_number = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, orderNumber).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresAccrualRepository) CreateOrder(ctx context.Context, orderNumber string, status model.Status) error {
	query := `
        INSERT INTO accrual_orders (order_number, status, accrual)
        VALUES ($1, $2, $3)
    `

	var dbStatus string
	switch status {
	case model.Registered:
		dbStatus = "REGISTERED"
	case model.Processing:
		dbStatus = "PROCESSING"
	case model.Invalid:
		dbStatus = "INVALID"
	case model.Processed:
		dbStatus = "PROCESSED"
	}

	_, err := r.db.ExecContext(ctx, query, orderNumber, dbStatus, nil)
	return err
}

func (r *PostgresAccrualRepository) UpdateOrderAccrual(ctx context.Context, orderNumber string, accrual float64, status model.Status) error {
	query := `
        UPDATE accrual_orders 
        SET status = $1, accrual = $2 
        WHERE order_number = $3
    `

	var dbStatus string
	switch status {
	case model.Registered:
		dbStatus = "REGISTERED"
	case model.Processing:
		dbStatus = "PROCESSING"
	case model.Invalid:
		dbStatus = "INVALID"
	case model.Processed:
		dbStatus = "PROCESSED"
	}

	_, err := r.db.ExecContext(ctx, query, dbStatus, accrual, orderNumber)
	return err
}

func (r *PostgresAccrualRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status model.Status) error {
	var dbStatus string
	switch status {
	case model.Registered:
		dbStatus = "REGISTERED"
	case model.Processing:
		dbStatus = "PROCESSING"
	case model.Invalid:
		dbStatus = "INVALID"
	case model.Processed:
		dbStatus = "PROCESSED"
	}

	query := `UPDATE accrual_orders SET status = $1 WHERE order_number = $2`
	_, err := r.db.ExecContext(ctx, query, dbStatus, orderNumber)
	return err
}

func (r *PostgresAccrualRepository) FindRewardByDescription(ctx context.Context, description string) (*model.GoodReward, error) {
	query := `
        SELECT description, reward_type, reward_value
        FROM goods_rewards
        WHERE description = $1
    `

	var reward model.GoodReward
	err := r.db.QueryRowContext(ctx, query, description).Scan(
		&reward.Description,
		&reward.RewardType,
		&reward.RewardValue,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}

	return &reward, nil
}

func (r *PostgresAccrualRepository) CreateGoodsReward(ctx context.Context, reward *model.GoodReward) error {
	dbRewardType := reward.RewardType
	if dbRewardType == "%" {
		dbRewardType = "percentage"
	} else if dbRewardType == "pt" {
		dbRewardType = "fixed"
	}

	query := `
        INSERT INTO goods_rewards (description, reward_type, reward_value, created_at)
        VALUES ($1, $2, $3, NOW())
    `

	_, err := r.db.ExecContext(ctx, query, reward.Description, dbRewardType, reward.RewardValue)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint" {
			return ErrDuplicateKey
		}
		return errors.New("failed to create goods reward")
	}

	return nil
}
