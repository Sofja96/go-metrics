package config

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/Sofja96/go-metrics.git/internal/utils"
)

// Config - структура хранения настроек сервера.
type Config struct {
	Address       string `env:"ADDRESS"`           // адрес и порт работы сервера
	StoreInterval int    `env:"STORE_INTERVAL"`    // интервал сохранения метрик
	FilePath      string `env:"FILE_STORAGE_PATH"` // путь к файлу хранилища
	Restore       bool   `env:"RESTORE"`           // указывает необходимость восстановить данные при старте сервера
	DatabaseDSN   string `env:"DATABASE_DSN"`      // строка подключения к БД
	HashKey       string `env:"KEY"`               // ключ аутентификации
	CryptoKey     string `env:"CRYPTO_KEY"`        // файл с приватным ключом сервера
	Config        string `env:"CONFIG"`            // файл настроки конфигурации
	TrustedSubnet string `env:"TRUSTED_SUBNET"`    // доверенная подсеть
}

const (
	DefaultAddress       = "localhost:8080"
	DefaultRestore       = true
	DefaultStoreInterval = 3
	DefaultFilePath      = "/tmp/metrics-db.json"
)

// TempConfig Временная структура для десериализации
type TempConfig struct {
	Address       string `json:"address"`
	StoreInterval string `json:"store_interval"`
	FilePath      string `json:"store_file"`
	Restore       bool   `json:"restore"`
	DatabaseDSN   string `json:"database_dsn,omitempty"`
	CryptoKey     string `json:"crypto_key,omitempty"`
	TrustedSubnet string `json:"trusted_subnet,omitempty"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Restore: DefaultRestore,
	}

	cfg.ParseFlags()

	if err := cfg.ApplyEnvVariables(); err != nil {
		log.Printf("WARNING: Failed to apply environment variables: %v", err)
	}

	if err := cfg.LoadFromFile(); err != nil {
		log.Printf("WARNING: Failed to load config file: %v", err)
	}

	if cfg.Address == "" {
		cfg.Address = DefaultAddress
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = DefaultStoreInterval
	}
	if cfg.FilePath == "" {
		cfg.FilePath = DefaultFilePath
	}

	return cfg, nil
}

// LoadFromFile функция для загрузки из файла и и применения конфигурации
func (cfg *Config) LoadFromFile() error {
	tempConfig, err := utils.ReadConfigFromFile[TempConfig](cfg.Config)
	if err != nil {
		return err
	}

	if tempConfig == nil {
		return nil
	}

	err = cfg.applyFileValues(tempConfig)
	if err != nil {
		return err
	}

	return nil
}

// applyFileValues функция для проверки установленных значений
func (cfg *Config) applyFileValues(tempConfig *TempConfig) error {

	if cfg.Address == "" && tempConfig.Address != "" {
		cfg.Address = tempConfig.Address
	}
	if cfg.StoreInterval == 0 && tempConfig.StoreInterval != "" {
		duration, err := time.ParseDuration(tempConfig.StoreInterval)
		if err != nil {
			return fmt.Errorf("invalid report_interval in config file: %w", err)
		}
		cfg.StoreInterval = int(duration.Seconds())
	}

	if cfg.FilePath == "" && tempConfig.FilePath != "" {
		cfg.FilePath = tempConfig.FilePath
	}

	if cfg.DatabaseDSN == "" && tempConfig.DatabaseDSN != "" {
		cfg.DatabaseDSN = tempConfig.DatabaseDSN
	}

	if cfg.CryptoKey == "" && tempConfig.CryptoKey != "" {
		cfg.CryptoKey = tempConfig.CryptoKey
	}

	if tempConfig.Restore != cfg.Restore {
		cfg.Restore = tempConfig.Restore
	}

	if cfg.TrustedSubnet == "" && tempConfig.TrustedSubnet != "" {
		cfg.TrustedSubnet = tempConfig.TrustedSubnet
	}

	return nil
}

// ApplyEnvVariables - применяет переменные окружения к конфигурации.
func (cfg *Config) ApplyEnvVariables() error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}

	return nil
}

// ParseFlags - функция настройки флагов и переменных окружения сервера.
func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run server")
	flag.StringVar(&cfg.FilePath, "f", cfg.FilePath, "path file storage to save data")
	flag.IntVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "interval for saving metrics on the server")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "need to load data at startup")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "connect to database")
	flag.StringVar(&cfg.HashKey, "k", cfg.HashKey, "key for hash")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path for public key file")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "Path to JSON config file")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "trusted subnet")

	flag.Parse()
}
