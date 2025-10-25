package gmstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// UserStorage methods
func (s *PostgresStorage) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var userID int64
	err := s.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return 0, gmmodel.ErrLoginAlreadyExists
			}
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (s *PostgresStorage) GetUserByLogin(ctx context.Context, login string) (*gmmodel.User, error) {
	query := `
		SELECT id, login, password_hash 
		FROM users 
		WHERE login = $1
	`

	user := &gmmodel.User{}

	err := s.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, gmmodel.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// OrderStorage methods
func (s *PostgresStorage) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	query := `
		INSERT INTO orders (user_id, order_number, status, uploaded_at)
		VALUES ($1, $2, $3, NOW())
	`

	_, err := s.db.ExecContext(ctx, query, userID, orderNumber, gmmodel.OrderStatusNew)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				// Check if owned by other user
				existingOrder, err := s.GetOrderByNumber(ctx, orderNumber)
				if err != nil {
					return fmt.Errorf("failed to check existing order: %w", err)
				}
				if existingOrder != nil && existingOrder.UserID != userID {
					return gmmodel.ErrOrderOwnedByOther
				}
				return gmmodel.ErrOrderAlreadyExists
			}
		}
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetOrderByNumber(ctx context.Context, orderNumber string) (*gmmodel.Order, error) {
	query := `
		SELECT id, user_id, order_number, status, accrual, uploaded_at
		FROM orders
		WHERE order_number = $1
	`

	order := &gmmodel.Order{}
	err := s.db.QueryRowContext(ctx, query, orderNumber).Scan(
		&order.ID,
		&order.UserID,
		&order.OrderNumber,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, gmmodel.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

func (s *PostgresStorage) GetUserOrders(ctx context.Context, userID int64) ([]*gmmodel.Order, error) {
	query := `
		SELECT id, user_id, order_number, status, accrual, uploaded_at
		FROM orders
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []*gmmodel.Order
	for rows.Next() {
		order := &gmmodel.Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderNumber,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

func (s *PostgresStorage) UpdateOrderStatus(ctx context.Context, orderNumber, status string, accrual *float64) error {
	query := `
		UPDATE orders
		SET status = $1, accrual = $2
		WHERE order_number = $3
	`

	var accrualParam interface{} = sql.NullFloat64{Float64: 0, Valid: false}
	if accrual != nil {
		accrualParam = *accrual
	}

	_, err := s.db.ExecContext(ctx, query, status, accrualParam, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetOrdersForProcessing(ctx context.Context) ([]*gmmodel.Order, error) {
	query := `
		SELECT id, user_id, order_number, status, accrual, uploaded_at
		FROM orders
		WHERE status IN ('NEW', 'PROCESSING', 'REGISTERED')
		FOR UPDATE SKIP LOCKED  -- To avoid concurrent processing
		LIMIT 10
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders for processing: %w", err)
	}
	defer rows.Close()

	var orders []*gmmodel.Order
	for rows.Next() {
		order := &gmmodel.Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderNumber,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// BalanceStorage methods
func (s *PostgresStorage) GetUserBalance(ctx context.Context, userID int64) (*gmmodel.Balance, error) {
	query := `
		SELECT 
			COALESCE(
				(SELECT SUM(accrual) 
				 FROM orders 
				 WHERE user_id = $1 AND status = $2 AND accrual IS NOT NULL), 
				0
			) - COALESCE(
				(SELECT SUM(sum) 
				 FROM withdrawals 
				 WHERE user_id = $1), 
				0
			) as current,
			COALESCE(
				(SELECT SUM(sum) 
				 FROM withdrawals 
				 WHERE user_id = $1), 
				0
			) as withdrawn
	`

	balance := &gmmodel.Balance{}
	err := s.db.QueryRowContext(ctx, query, userID, gmmodel.OrderStatusProcessed).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &gmmodel.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

// WithdrawalStorage methods
func (s *PostgresStorage) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check sufficient funds
	balance, err := s.getUserBalanceTx(tx, ctx, userID)
	if err != nil {
		return err
	}
	if balance.Current < sum {
		return gmmodel.ErrInsufficientFunds
	}

	// Insert withdrawal
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err = tx.ExecContext(ctx, query, userID, orderNumber, sum)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.New("order already withdrawn") // Or custom err
		}
		return fmt.Errorf("failed to create withdrawal: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetUserWithdrawals(ctx context.Context, userID int64) ([]*gmmodel.Withdrawal, error) {
	query := `
		SELECT order_number, sum, processed_at
		FROM withdrawals
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user withdrawals: %w", err)
	}
	defer rows.Close()

	var withdrawals []*gmmodel.Withdrawal
	for rows.Next() {
		withdrawal := &gmmodel.Withdrawal{}
		err := rows.Scan(
			&withdrawal.OrderNumber,
			&withdrawal.Sum,
			&withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return withdrawals, nil
}

// Helper for transaction (reuse balance query with tx)
func (s *PostgresStorage) getUserBalanceTx(tx *sql.Tx, ctx context.Context, userID int64) (*gmmodel.Balance, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN o.status = $2 THEN COALESCE(o.accrual, 0) ELSE 0 END), 0) AS current,
			COALESCE(SUM(w.sum), 0) AS withdrawn
		FROM orders o
		LEFT JOIN withdrawals w ON o.user_id = w.user_id
		WHERE o.user_id = $1
		GROUP BY o.user_id
	`

	balance := &gmmodel.Balance{}
	err := tx.QueryRowContext(ctx, query, userID, gmmodel.OrderStatusProcessed).Scan(
		&balance.Current,
		&balance.Withdrawn,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &gmmodel.Balance{Current: 0, Withdrawn: 0}, nil
		}
		return nil, fmt.Errorf("failed to get user balance in tx: %w", err)
	}

	return balance, nil
}

// Close method for cleanup
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
