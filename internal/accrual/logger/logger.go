package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = zap.Logger

func NewLogger(level zapcore.Level) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.DisableStacktrace = true
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	cfg.EncoderConfig.CallerKey = "logLine"

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger config: %w", err)
	}

	return logger, nil
}
