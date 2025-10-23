package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"

	"github.com/Popolzen/gofermat_team/internal/accrual/model"
)

func (h *accrualHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	number := chi.URLParam(r, "number")

	matched, _ := regexp.MatchString(`^\d+$`, number)
	if !matched {
		http.Error(w, model.ErrInvalidOrder.Error(), http.StatusBadRequest)
		return
	}

	accrual, err := h.accrualUseCase.GetOrderAccrual(r.Context(), number)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(accrual); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *accrualHandler) RegisterOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	var orderReq model.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if orderReq.Order == "" || len(orderReq.Goods) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	err := h.accrualUseCase.RegisterOrder(r.Context(), &orderReq)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrInvalidOrderFormat):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, model.ErrOrderAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *accrualHandler) RegisterGoodsReward(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rewardReq struct {
		Match      string  `json:"match"`
		Reward     float64 `json:"reward"`
		RewardType string  `json:"reward_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&rewardReq); err != nil {
		h.logger.Error("Failed to decode goods reward request", zap.Error(err))
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Received goods reward request",
		zap.String("match", rewardReq.Match),
		zap.Float64("reward", rewardReq.Reward),
		zap.String("reward_type", rewardReq.RewardType))

	reward := &model.GoodReward{
		Description: rewardReq.Match, 
		RewardValue: rewardReq.Reward,
		RewardType:  rewardReq.RewardType, 
	}

	if err := h.accrualUseCase.RegisterGoodsReward(r.Context(), reward); err != nil {
		h.logger.Error("Failed to register goods reward", 
			zap.String("match", rewardReq.Match),
			zap.Error(err))

		switch {
		case errors.Is(err, model.ErrRewardAlreadyExists):
			http.Error(w, "reward already exists", http.StatusConflict)
		case errors.Is(err, model.ErrInvalidRewardFormat):
			http.Error(w, "invalid reward format", http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "registered",
		"match":  rewardReq.Match,
	})
}
