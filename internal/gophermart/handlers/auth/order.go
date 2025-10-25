package gmhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	gmauth "github.com/Popolzen/gofermat_team/internal/gophermart/middleware/auth"
	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmservice "github.com/Popolzen/gofermat_team/internal/gophermart/service"
)

// UploadHandler обрабатывает загрузку нового номера заказа
func UploadHandler(orderService gmservice.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем Content-Type
		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
			return
		}

		// Получаем userID из контекста
		userID := gmauth.MustGetUserID(r.Context())

		// Читаем тело запроса
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		orderNumber := strings.TrimSpace(string(body))

		// Проверяем валидность номера заказа по алгоритму Луна
		if !isValidLuhn(orderNumber) {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}

		// Вызываем сервис для загрузки заказа
		err = orderService.Upload(r.Context(), userID, orderNumber)
		if err != nil {
			switch {
			case errors.Is(err, gmmodel.ErrOrderAlreadyExists):
				w.WriteHeader(http.StatusOK) // 200 для повторной загрузки тем же пользователем
			case errors.Is(err, gmmodel.ErrOrderOwnedByOther):
				http.Error(w, "Order already uploaded by another user", http.StatusConflict) // 409
			case errors.Is(err, gmmodel.ErrInvalidOrderNumber):
				http.Error(w, "Invalid order number", http.StatusUnprocessableEntity) // 422
			default:
				http.Error(w, "Failed to upload order", http.StatusInternalServerError) // 500
			}
			return
		}

		w.WriteHeader(http.StatusAccepted) // 202
	}
}

// GetOrdersHandler handles getting the list of user's orders
func GetOrdersHandler(orderService gmservice.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := gmauth.MustGetUserID(r.Context())

		orders, err := orderService.GetList(r.Context(), userID)
		if err != nil {
			if errors.Is(err, gmmodel.ErrOrderNotFound) { // 204
				w.WriteHeader(http.StatusNoContent) // 204
				return
			}
			http.Error(w, "Failed to get orders", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(orders); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func isValidLuhn(number string) bool {
	var sum int
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false // Номер должен содержать только цифры
		}
		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}
	return sum%10 == 0
}
