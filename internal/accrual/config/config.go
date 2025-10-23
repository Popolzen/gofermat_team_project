package config

import (
	"flag"
	"os"

	"go.uber.org/zap/zapcore"
)

type ServiceConfig struct {
	ServerPort      string
	RateLimitPerMin int
	RetryAfterSec   int
	DatabaseDSN     string
	LogLevel        zapcore.Level
}

func NewConfig() *ServiceConfig {
	cfg := &ServiceConfig{
		RateLimitPerMin: 60,
		RetryAfterSec:   60,
		LogLevel:        zapcore.InfoLevel,
	}

	var addrFlag, dsnFlag, logLevelFlag string
	flag.StringVar(&addrFlag, "a", ":8080", "Server address and port (e.g., :8080)")
	flag.StringVar(&dsnFlag, "d", "postgres://postgres:postgres@localhost:5432/loyalty?sslmode=disable", "Database DSN")
	flag.StringVar(&logLevelFlag, "l", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	cfg.ServerPort = getEnv("RUN_ADDRESS", addrFlag)
	cfg.DatabaseDSN = getEnv("DATABASE_URI", dsnFlag)

	switch getEnv("LOG_LEVEL", logLevelFlag) {
	case "debug":
		cfg.LogLevel = zapcore.DebugLevel
	case "info":
		cfg.LogLevel = zapcore.InfoLevel
	case "warn":
		cfg.LogLevel = zapcore.WarnLevel
	case "error":
		cfg.LogLevel = zapcore.ErrorLevel
	default:
		cfg.LogLevel = zapcore.InfoLevel
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
