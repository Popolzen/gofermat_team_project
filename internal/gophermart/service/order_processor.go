package gmservice

import (
	"context"
	"encoding/json"
	"fmt"
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

func (p *orderProcessor) Start(ctx context.Context) {
	go p.processLoop(ctx)
}

func (p *orderProcessor) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case <-p.done:
		return
	default:
	}
	close(p.done)
}

func (o orderProcessor) processLoop(ctx context.Context) {
	defer o.ticker.Stop()
	for {
		select {
		case <-o.ticker.C:
			fmt.Println("тикнули, берем заказы")
			o.processBatch(ctx)
		case <-ctx.Done():
			return
		case <-o.done:
			return
		}
	}

}

func (o *orderProcessor) processBatch(ctx context.Context) {

	orders, err := o.storage.GetOrdersForProcessing(ctx)
	if err != nil {
		log.Printf("Failed to get orders for processing: %v", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	// fmt.Println(orders)
	var wg sync.WaitGroup

	for _, order := range orders { // возможно ли много заказов?
		wg.Add(1)
		go func(order *gmmodel.Order) {
			defer wg.Done()

			o.processOrder(ctx, order)
		}(order)
	}
	wg.Wait()
}

func (o *orderProcessor) processOrder(ctx context.Context, order *gmmodel.Order) {
	// Процессируем заказ
	//я устал
	reqURL := fmt.Sprintf("%s/api/orders/%s", o.accrualURL, order.OrderNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)

	if err != nil {
		log.Printf("Failed to create request for order %s: %v", order.OrderNumber, err)
		return
	}
	resp, err := o.client.Do(req)
	if err != nil {
		log.Printf("Failed to process order %s: %v", order.OrderNumber, err)
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp struct {
			Status  string   `json:"status"`
			Accrual *float64 `json:"accrual"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			log.Printf("Failed to decode response for order %s: %v", order.OrderNumber, err)
			return
		}

		var newStatus string
		switch accrualResp.Status {
		case "REGISTERED":
			newStatus = gmmodel.OrderStatusRegistred
		case "PROCESSING":
			newStatus = gmmodel.OrderStatusProcessing
		case "INVALID":
			newStatus = gmmodel.OrderStatusInvalid
		case "PROCESSED":
			newStatus = gmmodel.OrderStatusProcessed
		default:
			log.Printf("Unknown status '%s' for order %s", accrualResp.Status, order.OrderNumber)
			return
		}

		if err := o.storage.UpdateOrderStatus(ctx, order.OrderNumber, newStatus, accrualResp.Accrual); err != nil {
			log.Printf("Failed to update order %s: %v", order.OrderNumber, err)
		} else {
			log.Printf("Updated order %s: %s (accrual: %v)", order.OrderNumber, newStatus, accrualResp.Accrual)
		}

	default:
		log.Printf("Unexpected status %d for order %s (retry next tick)", resp.StatusCode, order.OrderNumber)
	}
}
