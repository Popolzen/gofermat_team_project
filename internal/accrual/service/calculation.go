package service

import (
    "github.com/Popolzen/gofermat_team/internal/accrual/model"
)

type AccrualCalculator struct{}

func NewAccrualCalculator() *AccrualCalculator {
    return &AccrualCalculator{}
}

func (c *AccrualCalculator) CalculateReward(reward *model.GoodReward, price float64) float64 {
    switch reward.RewardType {
    case "percentage":
        return price * reward.RewardValue / 100
    case "fixed":
        return reward.RewardValue
    default:
        return 0
    }
}

func (c *AccrualCalculator) CalculateTotalAccrual(goods []model.Good, rewards map[string]*model.GoodReward) float64 {
    var totalAccrual float64

    for _, good := range goods {
        reward, exists := rewards[good.Description]
        if !exists {
            continue
		}

        totalAccrual += c.CalculateReward(reward, good.Price)
    }

    return totalAccrual
}
