package gmservice

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	gmauth "github.com/Popolzen/gofermat_team/internal/gophermart/middleware/auth"
	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmstorage "github.com/Popolzen/gofermat_team/internal/gophermart/storage"
)

type userService struct {
	storage gmstorage.UserStorage
	auth    *gmauth.Auth
}

func NewUserService(storage gmstorage.UserStorage, auth *gmauth.Auth) *userService {
	return &userService{storage: storage, auth: auth}
}

func (s *userService) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", fmt.Errorf("login and password required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := s.storage.CreateUser(ctx, login, string(hash))
	if err != nil {
		return "", err
	}

	token, err := gmauth.GenerateToken(id, login, s.auth.SecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

func (s *userService) Login(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", fmt.Errorf("login and password required")
	}

	user, err := s.storage.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, gmmodel.ErrUserNotFound) {
			return "", gmmodel.ErrUserNotFound
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", gmmodel.ErrUserNotFound
	}

	token, err := gmauth.GenerateToken(user.ID, user.Login, s.auth.SecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
