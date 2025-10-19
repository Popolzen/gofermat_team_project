package gmhandlers

import (
	"encoding/json"
	"errors"
	"net/http"

	gmauth "github.com/Popolzen/gofermat_team/internal/gophermart/middleware/auth"
	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmservice "github.com/Popolzen/gofermat_team/internal/gophermart/service"
)

// UploadOrderRequest represents the request body for uploading an order
type UploadOrderRequest struct {
	OrderNumber string `json:"order"`
}

// UploadHandler handles uploading a new order number
// Теперь без auth.AuthMiddleware внутри — оно применяется на роутере через .Use()
func UploadHandler(orderService gmservice.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// userID из ctx (AuthMiddleware уже добавил его)
		userID := gmauth.MustGetUserID(r.Context())

		var req UploadOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.OrderNumber == "" {
			http.Error(w, "Order number is required", http.StatusBadRequest)
			return
		}

		err := orderService.Upload(r.Context(), userID, req.OrderNumber)
		if err != nil {

			switch {
			case errors.Is(err, gmmodel.ErrOrderAlreadyExists): // 200 OK, если уже загружен пользователем (адаптируй имя ошибки, если ErrOrderAlreadyUploaded)
				w.WriteHeader(http.StatusOK)
			case errors.Is(err, gmmodel.ErrOrderOwnedByOther): // 409
				http.Error(w, "Order already uploaded by another user", http.StatusConflict)
			case errors.Is(err, gmmodel.ErrInvalidOrderNumber): // 422
				http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			default:
				http.Error(w, "Failed to upload order", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusAccepted) // 202
	}
}

// GetOrdersHandler handles getting the list of user's orders
// Аналогично, без middleware
func GetOrdersHandler(orderService gmservice.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := gmauth.MustGetUserID(r.Context())

		orders, err := orderService.GetList(r.Context(), userID)
		if err != nil {
			if errors.Is(err, gmmodel.ErrOrderNotFound) { // Или ErrNoOrders, если добавишь в model; для 204
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
