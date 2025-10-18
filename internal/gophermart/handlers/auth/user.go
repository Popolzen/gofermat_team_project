package gmhandlers

import (
	"encoding/json"
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
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "content-type must be application/json", http.StatusBadRequest)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

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
func LoginHandler(userService gmservice.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// реализация Login
		// использует userService
	}
}
