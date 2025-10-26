package gmhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	gmauth "github.com/Popolzen/gofermat_team/internal/gophermart/middleware/auth"
	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmservice "github.com/Popolzen/gofermat_team/internal/gophermart/service"
)

// GetBalanceHandler возвращает текущий баланс пользователя
func GetBalanceHandler(balanceService gmservice.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем userID из контекста
		userID := gmauth.MustGetUserID(r.Context())

		// Получаем баланс
		balance, err := balanceService.Get(r.Context(), userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Возвращаем JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(balance)
	}
}

// WithdrawRequest структура запроса на списание
type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// WithdrawHandler обрабатывает списание баллов
func WithdrawHandler(balanceService gmservice.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверка Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "content-type must be application/json", http.StatusBadRequest)
			return
		}

		// Получаем userID из контекста
		userID := gmauth.MustGetUserID(r.Context())

		// Парсим запрос
		var req WithdrawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

		// Валидация полей
		if req.Order == "" || req.Sum <= 0 {
			http.Error(w, "order and sum are required", http.StatusBadRequest)
			return
		}

		// Вызываем сервис
		err := balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
		if err != nil {
			// Обработка специфичных ошибок
			if errors.Is(err, gmmodel.ErrInsufficientFunds) {
				http.Error(w, "insufficient funds", http.StatusPaymentRequired) // 402
				return
			}
			if errors.Is(err, gmmodel.ErrInvalidOrderNumber) {
				http.Error(w, "invalid order number", http.StatusUnprocessableEntity) // 422
				return
			}
			// Общая ошибка
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Успешное списание
		w.WriteHeader(http.StatusOK)
	}
}

// GetWithdrawalsHandler возвращает историю списаний
func GetWithdrawalsHandler(balanceService gmservice.BalanceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем userID из контекста
		userID := gmauth.MustGetUserID(r.Context())

		// Получаем историю списаний
		withdrawals, err := balanceService.GetWithdrawals(r.Context(), userID)
		if err != nil {
			// Если нет списаний
			if errors.Is(err, gmmodel.ErrNoWithdrawals) {
				w.WriteHeader(http.StatusNoContent) // 204
				return
			}
			// Общая ошибка
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Возвращаем JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(withdrawals)
	}
}
