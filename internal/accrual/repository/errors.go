package repository

import (
	"database/sql"
	"errors"

	"github.com/Popolzen/gofermat_team/internal/accrual/model"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = model.ErrNotFound
)

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows || err == ErrNotFound
}

func IsDuplicateKeyError(err error) bool {
    return errors.Is(err, ErrDuplicateKey)
}
