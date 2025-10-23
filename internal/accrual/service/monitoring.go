package service

import (
    "time"
    
    "go.uber.org/zap"
    "github.com/Popolzen/gofermat_team/internal/accrual/logger"
)

type MonitoringService struct {
    processor *OrderProcessor
    logger    *logger.Logger
    stopChan  chan struct{}
}

func NewMonitoringService(processor *OrderProcessor, logger *logger.Logger) *MonitoringService {
    return &MonitoringService{
        processor: processor,
        logger:    logger,
        stopChan:  make(chan struct{}),
    }
}

func (m *MonitoringService) StartMonitoring() {
    go m.monitorQueueSize()
    go m.monitorResults()
    m.logger.Info("Monitoring service started")
}

func (m *MonitoringService) Stop() {
    close(m.stopChan)
    m.logger.Info("Monitoring service stopped")
}

func (m *MonitoringService) monitorQueueSize() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            stats := m.processor.GetStats()
            queueLen := stats["queue_len"]
            
            if queueLen > 50 {
                m.logger.Warn("High queue load", 
                    zap.Int("queue_length", queueLen))
            } else if queueLen > 0 {
                m.logger.Debug("Queue monitoring", 
                    zap.Int("queue_length", queueLen))
            }
            
        case <-m.stopChan:
            m.logger.Info("Queue monitoring stopped")
            return
        }
    }
}

func (m *MonitoringService) monitorResults() {
    results := m.processor.GetResults()
    
    for {
        select {
        case result, ok := <-results:
            if !ok {
                m.logger.Info("Results monitoring stopped")
                return
            }
            
            if result.Error != nil {
                m.logger.Warn("Processing error in monitoring", 
                    zap.String("order", result.OrderNumber), 
                    zap.Error(result.Error))
            } else {
                m.logger.Debug("Processing success in monitoring", 
                    zap.String("order", result.OrderNumber), 
                    zap.Float64("accrual", result.Accrual))
            }
            
        case <-m.stopChan:
            m.logger.Info("Results monitoring stopped")
            return
        }
    }
}
