package gmhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmservice "github.com/Popolzen/gofermat_team/internal/gophermart/service"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterHandler(userService gmservice.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Проверка Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "content-type must be application/json", http.StatusBadRequest)
			return
		}

		// Парсинг запроса
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

		// Валидация полей
		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password are required", http.StatusBadRequest)
			return
		}

		token, err := userService.Register(r.Context(), req.Login, req.Password)
		if err != nil {
			switch err {
			case gmmodel.ErrLoginAlreadyExists:
				http.Error(w, "login already exists", http.StatusConflict)
			default:
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Автоматическая аутентификация: возвращаем token в header
		w.Header().Set("Authorization", "Bearer "+token)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Опционально: пустой JSON body для 200, или echo request
		json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
	}
}

// LoginHandler обрабатывает аутентификацию пользователя
func LoginHandler(userService gmservice.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверка Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "content-type must be application/json", http.StatusBadRequest)
			return
		}

		// Парсинг запроса
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

		// Валидация полей
		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password are required", http.StatusBadRequest)
			return
		}

		// Вызов сервиса для аутентификации
		token, err := userService.Login(r.Context(), req.Login, req.Password)
		if err != nil {
			// Обработка специфичных ошибок
			if errors.Is(err, gmmodel.ErrUserNotFound) ||
				errors.Is(err, gmmodel.ErrInvalidPassword) {
				http.Error(w, "invalid credentials", http.StatusUnauthorized)
				return
			}
			// Общая ошибка сервера
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Возвращаем токен в заголовке
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}
