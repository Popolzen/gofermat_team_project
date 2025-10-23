package gmservice

import (
	"context"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmstorage "github.com/Popolzen/gofermat_team/internal/gophermart/storage"
)

type balanceService struct {
	balanceStorage    gmstorage.BalanceStorage
	withdrawalStorage gmstorage.WithdrawalStorage
}

func NewBalanceService(balanceStorage gmstorage.BalanceStorage, withdrawalStorage gmstorage.WithdrawalStorage) *balanceService {
	return &balanceService{balanceStorage: balanceStorage, withdrawalStorage: withdrawalStorage}
}

// type BalanceService interface {
// 	Get(ctx context.Context, userID int64) (*BalanceDTO, error)
// 	Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error
// 	GetWithdrawals(ctx context.Context, userID int64) ([]*WithdrawalDTO, error)
// }

func (b balanceService) Get(ctx context.Context, userID int64) (*BalanceDTO, error) {

	balance, err := b.balanceStorage.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &BalanceDTO{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}, nil
}

// Withdraw списывает баллы со счёта пользователя
func (b *balanceService) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	// Валидация суммы
	if sum <= 0 {
		return gmmodel.ErrInvalidWithdrawAmount
	}

	// Проверка номера заказа алгоритмом Луна
	// if !utils.ValidateOrderNumber(orderNumber) {
	// 	return gmmodel.ErrInvalidOrderNumber
	// }

	// Получаем текущий баланс
	balance, err := b.balanceStorage.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}

	//  Проверяем достаточно ли средств
	if balance.Current < sum {
		return gmmodel.ErrInsufficientFunds
	}

	//  Создаём списание
	err = b.withdrawalStorage.CreateWithdrawal(ctx, userID, orderNumber, sum)
	if err != nil {
		return err
	}

	return nil
}

// GetWithdrawals возвращает историю списаний пользователя
func (b *balanceService) GetWithdrawals(ctx context.Context, userID int64) ([]*WithdrawalDTO, error) {
	// Получаем списания из storage
	withdrawals, err := b.withdrawalStorage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Если списаний нет
	if len(withdrawals) == 0 {
		return nil, gmmodel.ErrNoWithdrawals
	}

	// Конвертируем в DTO
	result := make([]*WithdrawalDTO, 0, len(withdrawals))
	for _, w := range withdrawals {
		result = append(result, &WithdrawalDTO{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}

	return result, nil
}
