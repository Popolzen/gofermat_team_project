package gmdatabase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	"github.com/jackc/pgx/v5/pgconn"
)

type Database struct {
	DB *sql.DB
}

func (d *Database) CreateUser(ctx context.Context, login, passwordHash string) (int64, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var userID int64
	err := d.DB.QueryRowContext(ctx, query, login, passwordHash).Scan(&userID)
	if err != nil {
		// Проверяем на дубликат логина
		if pgErr, ok := err.(*pgconn.PgError); ok {
			// 23505 - unique_violation
			if pgErr.Code == "23505" {
				return 0, gmmodel.ErrLoginAlreadyExists
			}
		}
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (d *Database) GetUserByLogin(ctx context.Context, login string) (*gmmodel.User, error) {

	query := `
	SELECT * From users WHERE login = $1`

	user := &gmmodel.User{}

	err := d.DB.QueryRowContext(ctx, query, login).Scan(
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
