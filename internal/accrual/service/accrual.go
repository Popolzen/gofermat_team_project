package service

import (
	"context"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/Popolzen/gofermat_team/internal/accrual/logger"
	"github.com/Popolzen/gofermat_team/internal/accrual/model"
	"github.com/Popolzen/gofermat_team/internal/accrual/repository"
)

type AccrualUseCaseImpl struct {
	repo       repository.AccrualRepository
	processor  *OrderProcessor
	monitoring *MonitoringService
	calculator *AccrualCalculator
	logger     *logger.Logger
}

func NewAccrualUseCase(repo repository.AccrualRepository, logger *logger.Logger) AccrualUseCase {
	calculator := NewAccrualCalculator()
	processor := NewOrderProcessor(repo, logger, 3)
	monitoring := NewMonitoringService(processor, logger)
	monitoring.StartMonitoring()

	return &AccrualUseCaseImpl{
		repo:       repo,
		processor:  processor,
		monitoring: monitoring,
		calculator: calculator,
		logger:     logger,
	}
}

func (u *AccrualUseCaseImpl) GetOrderAccrual(ctx context.Context, order string) (*model.OrderAccrual, error) {
	if !isLuhnValid(order) {
		return &model.OrderAccrual{
			Order:  order,
			Status: model.Invalid,
		}, nil
	}

	accrual, err := u.repo.GetOrderAccrual(ctx, order)
	if err != nil {
		return nil, err
	}
	return accrual, nil
}

func (u *AccrualUseCaseImpl) RegisterOrder(ctx context.Context, orderReq *model.OrderRequest) error {
	if !isValidOrderNumber(orderReq.Order) {
		return model.ErrInvalidOrderFormat
	}

	if len(orderReq.Goods) == 0 {
		return model.ErrInvalidOrderFormat
	}

	for _, good := range orderReq.Goods {
		if good.Description == "" || good.Price <= 0 {
			return model.ErrInvalidOrderFormat
		}
	}

	exists, err := u.repo.OrderExists(ctx, orderReq.Order)
	if err != nil {
	return err
	}
	if exists {
		return model.ErrOrderAlreadyExists
	}

	err = u.repo.CreateOrder(ctx, orderReq.Order, model.Registered)
	if err != nil {
		return err
	}

	err = u.processor.SubmitOrder(ctx, orderReq)
	if err != nil {
		u.logger.Error("Failed to submit order for processing",
			zap.String("order", orderReq.Order),
			zap.Error(err))
		return err
	}

	u.logger.Info("Order registered and queued for processing",
		zap.String("order", orderReq.Order))
	return nil
}

func (u *AccrualUseCaseImpl) CalculateOrderAccrual(ctx context.Context, goods []model.Good) (float64, error) {
	if len(goods) == 0 {
		return 0, model.ErrInvalidOrderFormat
	}

	rewards := make(map[string]*model.GoodReward)
	for _, good := range goods {
		reward, err := u.repo.FindRewardByDescription(ctx, good.Description)
		if err != nil && !repository.IsNotFound(err) {
			return 0, err
		}
		if reward != nil {
			rewards[good.Description] = reward
		}
	}

	return u.calculator.CalculateTotalAccrual(goods, rewards), nil
}

func (u *AccrualUseCaseImpl) Stop() {
	u.logger.Info("Stopping accrual service...")
	u.monitoring.Stop()
	u.processor.Stop()
	u.logger.Info("Accrual service stopped")
}

func isLuhnValid(number string) bool {
	digits := make([]int, len(number))
	for i, char := range number {
		d, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}
		digits[i] = d
	}

	for i := len(digits) - 2; i >= 0; i -= 2 {
		doubled := digits[i] * 2
		if doubled > 9 {
			doubled -= 9
		}
		digits[i] = doubled
	}

	sum := 0
	for _, d := range digits {
		sum += d
	}

	return sum%10 == 0
}

func isValidOrderNumber(number string) bool {
	return isLuhnValid(number)
}

func (u *AccrualUseCaseImpl) RegisterGoodsReward(ctx context.Context, reward *model.GoodReward) error {
	u.logger.Debug("Validating goods reward",
		zap.String("description", reward.Description),
		zap.String("reward_type", reward.RewardType),
		zap.Float64("reward_value", reward.RewardValue))

	if err := u.validateGoodsReward(reward); err != nil {
		u.logger.Error("Goods reward validation failed",
			zap.String("description", reward.Description),
			zap.String("reward_type", reward.RewardType),
			zap.Float64("reward_value", reward.RewardValue),
			zap.Error(err))
		return err
	}

	err := u.repo.CreateGoodsReward(ctx, reward)
	if err != nil {
		if repository.IsDuplicateKeyError(err) {
			return model.ErrRewardAlreadyExists
		}
		return err
	}

	u.logger.Info("Goods reward registered successfully",
		zap.String("description", reward.Description),
		zap.String("reward_type", reward.RewardType),
		zap.Float64("reward_value", reward.RewardValue))
	return nil
}

func (u *AccrualUseCaseImpl) validateGoodsReward(reward *model.GoodReward) error {
	if strings.TrimSpace(reward.Description) == "" {
		u.logger.Debug("Validation failed: empty description")
		return model.ErrInvalidRewardFormat
	}

	if reward.RewardValue <= 0 {
		u.logger.Debug("Validation failed: reward value <= 0", zap.Float64("reward_value", reward.RewardValue))
		return model.ErrInvalidRewardFormat
	}

	validTypes := map[string]bool{
		"percentage": true,
		"fixed":      true,
		"%":          true,  
		"pt":         true,  
	}

	if !validTypes[reward.RewardType] {
		u.logger.Debug("Validation failed: invalid reward type", zap.String("reward_type", reward.RewardType))
		return model.ErrInvalidRewardFormat
	}

	return nil
}
