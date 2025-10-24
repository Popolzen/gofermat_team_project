package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Popolzen/gofermat_team/internal/accrual/logger"
	"github.com/Popolzen/gofermat_team/internal/accrual/model"
	"github.com/Popolzen/gofermat_team/internal/accrual/repository"
	"go.uber.org/zap"
)

type OrderProcessor struct {
	repo       repository.AccrualRepository
	calculator *AccrualCalculator
	logger     *logger.Logger
	jobQueue   chan *processJob
	resultChan chan *model.ProcessingResult
	workers    int
	wg         sync.WaitGroup
	stopChan   chan struct{}
	isStopped  bool
	mu         sync.RWMutex
}

type processJob struct {
	orderReq *model.OrderRequest
	ctx      context.Context
}

func NewOrderProcessor(repo repository.AccrualRepository, logger *logger.Logger, workers int) *OrderProcessor {
	calculator := NewAccrualCalculator()

	processor := &OrderProcessor{
		repo:       repo,
		calculator: calculator,
		logger:     logger,
		jobQueue:   make(chan *processJob, 100),
		resultChan: make(chan *model.ProcessingResult, 100),
		workers:    workers,
		stopChan:   make(chan struct{}),
	}

	for i := range workers {
		processor.wg.Add(1)
		go processor.worker(i)
	}

	processor.wg.Add(1)
	go processor.resultAggregator()

	processor.logger.Info("Order processor started", zap.Int("workers", workers))
	return processor
}

func (p *OrderProcessor) SubmitOrder(ctx context.Context, orderReq *model.OrderRequest) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.isStopped {
		p.logger.Error("Processor is stopped, cannot submit order")
		return errors.New("processor is stopped")
	}

	reqCopy := &model.OrderRequest{
		Order: orderReq.Order,
		Goods: make([]model.Good, len(orderReq.Goods)),
	}
	copy(reqCopy.Goods, orderReq.Goods)

	job := &processJob{
		orderReq: reqCopy,
		ctx: context.Background(), 	}

	select {
	case p.jobQueue <- job:
		p.logger.Info("Order submitted to job queue",
			zap.String("order", orderReq.Order),
			zap.Int("queue_size", len(p.jobQueue)))
		return nil
	case <-p.stopChan:
		p.logger.Error("Stop channel closed while submitting order")
		return errors.New("processor is stopped")
	default:
		p.logger.Warn("Order queue is full",
			zap.String("order", orderReq.Order),
			zap.Int("queue_capacity", cap(p.jobQueue)))
		return errors.New("order queue is full")
	}
}

func (p *OrderProcessor) worker(id int) {
	defer p.wg.Done()

	p.logger.Info("Worker started", zap.Int("worker", id))

	for {
		select {
		case job, ok := <-p.jobQueue:
			if !ok {
				p.logger.Info("Job queue closed, worker stopping", zap.Int("worker", id))
				return
			}
			p.logger.Info("Worker received job from queue",
				zap.Int("worker", id),
				zap.String("order", job.orderReq.Order))
			p.processOrder(job.ctx, id, job.orderReq)

		case <-p.stopChan:
			p.logger.Info("Worker received stop signal", zap.Int("worker", id))
			return
		}
	}
}

func (p *OrderProcessor) processOrder(ctx context.Context, workerID int, orderReq *model.OrderRequest) {
	p.logger.Info("Worker started processing order",
		zap.Int("worker", workerID),
		zap.String("order", orderReq.Order))

	err := p.repo.UpdateOrderStatus(ctx, orderReq.Order, model.Processing)
	if err != nil {
		p.logger.Error("Worker error updating status to PROCESSING",
			zap.Int("worker", workerID),
			zap.String("order", orderReq.Order),
			zap.Error(err))

		p.resultChan <- &model.ProcessingResult{
			OrderNumber: orderReq.Order,
			Status:      model.Invalid,
			Error:       err,
		}
		return
	}

	time.Sleep(2 * time.Second)

	accrual, err := p.calculateAccrual(ctx, orderReq.Goods)
	if err != nil {
		p.logger.Error("Worker error calculating accrual",
			zap.Int("worker", workerID),
			zap.String("order", orderReq.Order),
			zap.Error(err))

		p.repo.UpdateOrderStatus(ctx, orderReq.Order, model.Invalid)

		p.resultChan <- &model.ProcessingResult{
			OrderNumber: orderReq.Order,
			Status:      model.Invalid,
			Error:       err,
		}
		return
	}

	err = p.repo.UpdateOrderAccrual(ctx, orderReq.Order, accrual, model.Processed)
	if err != nil {
		p.logger.Error("Worker error updating order accrual",
			zap.Int("worker", workerID),
			zap.String("order", orderReq.Order),
			zap.Error(err))

		p.resultChan <- &model.ProcessingResult{
			OrderNumber: orderReq.Order,
			Accrual:     accrual,
			Status:      model.Processed,
			Error:       err,
		}
		return
	}

	p.logger.Info("Worker finished processing order",
		zap.Int("worker", workerID),
		zap.String("order", orderReq.Order),
		zap.Float64("accrual", accrual))

	p.resultChan <- &model.ProcessingResult{
		OrderNumber: orderReq.Order,
		Accrual:     accrual,
		Status:      model.Processed,
		Error:       nil,
	}
}

func (p *OrderProcessor) calculateAccrual(ctx context.Context, goods []model.Good) (float64, error) {
	descriptions := make([]string, 0, len(goods))
	for _, good := range goods {
		descriptions = append(descriptions, good.Description)
	}

	rewards := make(map[string]*model.GoodReward)
	for _, description := range descriptions {
		reward, err := p.repo.FindRewardByDescription(ctx, description)
		if err != nil {
			if repository.IsNotFound(err) {
				continue
			}
			return 0, err
		}
		rewards[description] = reward
	}

	return p.calculator.CalculateTotalAccrual(goods, rewards), nil
}

func (p *OrderProcessor) resultAggregator() {
	defer p.wg.Done()

	p.logger.Info("Result aggregator started")

	stats := struct {
		processed int
		failed    int
		total     int
	}{}

	for {
		select {
		case result, ok := <-p.resultChan:
			if !ok {
				p.logger.Info("Result aggregator stopping",
					zap.Int("processed", stats.processed),
					zap.Int("failed", stats.failed),
					zap.Int("total", stats.total))
				return
			}

			stats.total++
			if result.Error != nil {
				stats.failed++
				p.logger.Error("Order processing failed",
					zap.String("order", result.OrderNumber),
					zap.Error(result.Error))
			} else {
				stats.processed++
				p.logger.Info("Order processing completed",
					zap.String("order", result.OrderNumber),
					zap.Float64("accrual", result.Accrual))
			}

		case <-p.stopChan:
			p.logger.Info("Result aggregator received stop signal",
				zap.Int("processed", stats.processed),
				zap.Int("failed", stats.failed),
				zap.Int("total", stats.total))
			return
		}
	}
}

func (p *OrderProcessor) GetResults() <-chan *model.ProcessingResult {
	return p.resultChan
}

func (p *OrderProcessor) GetStats() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]int{
		"workers":    p.workers,
		"queue_len":  len(p.jobQueue),
		"is_stopped": 0,
	}
}

func (p *OrderProcessor) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isStopped {
		return
	}

	p.logger.Info("Stopping order processor...")
	p.isStopped = true

	close(p.stopChan)
	close(p.jobQueue)
	close(p.resultChan)

	p.wg.Wait()
	p.logger.Info("Order processor stopped")
}
