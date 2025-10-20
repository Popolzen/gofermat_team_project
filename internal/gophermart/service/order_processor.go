package gmservice

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	gmmodel "github.com/Popolzen/gofermat_team/internal/gophermart/model"
	gmstorage "github.com/Popolzen/gofermat_team/internal/gophermart/storage"
)

type orderProcessor struct {
	storage    gmstorage.OrderStorage
	accrualURL string
	client     *http.Client
	ticker     *time.Ticker
	done       chan struct{}
	mu         sync.Mutex
}

func NewOrderProcessor(storage gmstorage.OrderStorage, accrualURL string) OrderProcessor {
	return &orderProcessor{
		storage:    storage,
		accrualURL: accrualURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		ticker: time.NewTicker(10 * time.Second), // каждые 10 сек
		done:   make(chan struct{}),
	}
}

func (o orderProcessor) Start(ctx context.Context) {

}

func (o orderProcessor) Stop() {

}

func (o orderProcessor) processLoop(ctx context.Context) {
	defer o.ticker.Stop()

	select {
	case <-o.ticker.C:
		o.processBatch(ctx)
	case <-ctx.Done():

	case <-o.done:
		return
	}
}

func (o orderProcessor) processBatch(ctx context.Context) {

	orders, err := o.storage.GetOrdersForProcessing(ctx)
	if err != nil {
		log.Printf("Failed to get orders for processing: %v", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	var wg sync.WaitGroup

	for _, order := range orders {
		wg.Add(1)
		go func(order *gmmodel.Order) {
			defer wg.Done()

			o.processOrder(ctx, order)
		}(order)
	}
	wg.Wait()
}

func (o orderProcessor) processOrder(ctx context.Context, order *gmmodel.Order) {
	// Процессируем заказ
	//я устал
}
