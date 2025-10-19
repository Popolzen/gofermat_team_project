package gmservice

import (
	"context"
	"net/http"
	"sync"
	"time"

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
