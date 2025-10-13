package gmconfig

import (
	"flag"
	"log"
	"os"
)

const (
	DefaultServerAddr = "localhost:8080"
)

type Config struct {
	ServerAddr   string `env:"RUN_ADDRESS"`
	DBurl        string `env:"DATABASE_URI"`
	AcServerAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() *Config {
	c := &Config{
		ServerAddr: DefaultServerAddr,
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
}

func (c *Config) parseFlags() {
	var (
		serverAddr   string
		dbURL        string
		acServerAddr string
	)

	flag.StringVar(&serverAddr, "a", "", "server address (host:port)")
	flag.StringVar(&dbURL, "d", "", "database connection URL")
	flag.StringVar(&acServerAddr, "r", "", "accrual system address")
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
}

func (c *Config) Validate() error {
	if c.DBurl == "" {
		log.Fatal("DATABASE_URI is required")
	}
	if c.AcServerAddr == "" {
		log.Fatal("ACCRUAL_SYSTEM_ADDRESS is required")
	}
	return nil
}
