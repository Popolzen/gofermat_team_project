package gmconfig

import (
	"errors"
	"flag"
	"os"
)

const (
	DefaultServerAddr   = "localhost:8080"
	DefaultSecretKey    = "jwt-key"
	DefaultAcServerAddr = "http://localhost:8081"
)

type Config struct {
	ServerAddr   string `env:"RUN_ADDRESS"`
	DBurl        string `env:"DATABASE_URI"`
	AcServerAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey    string `env:"SECRET_KEY"`
}

func NewConfig() *Config {
	c := &Config{
		ServerAddr:   DefaultServerAddr,
		SecretKey:    DefaultSecretKey,
		AcServerAddr: DefaultAcServerAddr,
	}

	// 1. Загружаем ENV переменные
	c.parseEnv()

	// 2. Загружаем флаги (наивысший приоритет)
	c.parseFlags()

	return c
}

func (c *Config) parseEnv() {
	if addr, exists := os.LookupEnv("RUN_ADDRESS"); exists {
		c.ServerAddr = addr
	}
	if dbURL, exists := os.LookupEnv("DATABASE_URI"); exists {
		c.DBurl = dbURL
	}
	if acAddr, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); exists {
		c.AcServerAddr = acAddr
	}
	if secretKey, exists := os.LookupEnv("SECRET_KEY"); exists {
		c.SecretKey = secretKey
	}
}

func (c *Config) parseFlags() {
	var (
		serverAddr   string
		dbURL        string
		acServerAddr string
		secretKey    string
	)

	flag.StringVar(&serverAddr, "a", "", "server address (host:port)")
	flag.StringVar(&dbURL, "d", "", "database connection URL")
	flag.StringVar(&acServerAddr, "r", "", "accrual system address")
	flag.StringVar(&secretKey, "s", "", "JWT secret key")
	flag.Parse()

	// Применяем флаги только если они были явно указаны
	if serverAddr != "" {
		c.ServerAddr = serverAddr
	}
	if dbURL != "" {
		c.DBurl = dbURL
	}
	if acServerAddr != "" {
		c.AcServerAddr = acServerAddr
	}
	if secretKey != "" {
		c.SecretKey = secretKey
	}
}

func (c *Config) Validate() error {
	if c.DBurl == "" {
		return errors.New("DATABASE_URI is required (set via -d flag or DATABASE_URI env)")
	}
	if c.AcServerAddr == "" {
		return errors.New("ACCRUAL_SYSTEM_ADDRESS is required (set via -r flag or ACCRUAL_SYSTEM_ADDRESS env)")
	}
	if len(c.SecretKey) < 32 {
		return errors.New("SECRET_KEY must be at least 32 characters long")
	}
	return nil
}
