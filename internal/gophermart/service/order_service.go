package gmservice

import (
	"context"
	"errors"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmstorage "github.com/Popolzen/gofermat_team/internal/gophermart/storage"
	// "github.com/Popolzen/gofermat_team/internal/gophermart/utils"
)

// orderService реализация OrderService
type orderService struct {
	storage gmstorage.OrderStorage
}

// NewOrderService создает новый экземпляр OrderService
func NewOrderService(storage gmstorage.OrderStorage) OrderService {
	return &orderService{
		storage: storage,
	}
}

// Upload загружает новый номер заказа для пользователя
func (s *orderService) Upload(ctx context.Context, userID int64, orderNumber string) error {
	// Проверяем номер заказа алгоритмом Луна
	// if !utils.ValidateOrderNumber(orderNumber) {
	// 	return gmmodel.ErrInvalidOrderNumber
	// }

	// Проверяем существует ли уже такой заказ
	existingOrder, err := s.storage.GetOrderByNumber(ctx, orderNumber)
	if err != nil && !errors.Is(err, gmmodel.ErrOrderNotFound) {
		// Ошибка БД
		return err
	}

	// 3. Если заказ существует
	if existingOrder != nil {
		// Проверяем кто загрузил
		if existingOrder.UserID == userID {
			// Тот же пользователь загружает повторно
			return gmmodel.ErrOrderAlreadyExists // 200 OK
		} else {
			// Другой пользователь уже загрузил
			return gmmodel.ErrOrderOwnedByOther // 409 Conflict
		}
	}

	// Создаём новый заказ со статусом NEW
	err = s.storage.CreateOrder(ctx, userID, orderNumber)
	if err != nil {
		return err
	}

	// Возвращаем nil = успех (202 Accepted)
	return nil
}

// GetList возвращает список заказов пользователя
func (s *orderService) GetList(ctx context.Context, userID int64) ([]*OrderDTO, error) {
	// Получаем заказы из БД
	orders, err := s.storage.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Если заказов нет
	if len(orders) == 0 {
		return nil, gmmodel.ErrNoOrders //  204 No Content
	}

	// Конвертируем в DTO
	result := make([]*OrderDTO, 0, len(orders))
	for _, order := range orders {
		result = append(result, &OrderDTO{
			Number:     order.OrderNumber,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	return result, nil
}
