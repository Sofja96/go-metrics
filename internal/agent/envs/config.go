package envs

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/Sofja96/go-metrics.git/internal/utils"
)

// Config - струтура хранения настроек агента.
type Config struct {
	Address        string `env:"ADDRESS"`         // адрес работы агента сбора метрик
	GrpcAddress    string `env:"GRPC_ADDRESS"`    // адрес работы grpc агента сбора метрик
	ReportInterval int    `env:"REPORT_INTERVAL"` // интервал отправки метрик
	PollInterval   int    `env:"POLL_INTERVAL"`   // интервал сбора метрик
	HashKey        string `env:"KEY"`             // ключ аутентификации
	RateLimit      int    `env:"RATE_LIMIT"`      // ограничение на количество исходящих запросов
	CryptoKey      string `env:"CRYPTO_KEY"`      // файл с публичным ключом сервера
	Config         string `env:"CONFIG"`          // файл настроки конфигурации
	UseGRPC        bool   `env:"USE_GRPC"`        // флаг включения grpc
}

const (
	DefaultAddress        = "localhost:8080"
	DefaultReportInterval = 10
	DefaultPollInterval   = 2
	DefaultRateLimit      = 1
	DefaultUseGRPC        = false
)

// TempConfig Временная структура для десериализации
type TempConfig struct {
	Address        string `json:"address"`
	GrpcAddress    string `json:"grpc_address"`
	PollInterval   string `json:"poll_interval"`
	ReportInterval string `json:"report_interval"`
	CryptoKey      string `json:"crypto_key"`
	UseGRPC        bool   `json:"use_grpc"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		UseGRPC: DefaultUseGRPC,
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
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = DefaultReportInterval
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = DefaultPollInterval
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = DefaultRateLimit
	}

	return cfg, nil
}

// applyFileValues функция для проверки установленных значений
func (cfg *Config) applyFileValues(tempConfig *TempConfig) error {

	if cfg.Address == "" && tempConfig.Address != "" {
		cfg.Address = tempConfig.Address
	}
	if cfg.GrpcAddress == "" && tempConfig.GrpcAddress != "" {
		cfg.GrpcAddress = tempConfig.GrpcAddress
	}
	if cfg.ReportInterval == 0 && tempConfig.ReportInterval != "" {
		duration, err := time.ParseDuration(tempConfig.ReportInterval)
		if err != nil {
			return fmt.Errorf("invalid report_interval in config file: %w", err)
		}
		cfg.ReportInterval = int(duration.Seconds())
	}
	if cfg.PollInterval == 0 && tempConfig.PollInterval != "" {
		duration, err := time.ParseDuration(tempConfig.PollInterval)
		if err != nil {
			return fmt.Errorf("invalid poll_interval in config file: %w", err)
		}
		cfg.PollInterval = int(duration.Seconds())
	}
	if cfg.CryptoKey == "" && tempConfig.CryptoKey != "" {
		cfg.CryptoKey = tempConfig.CryptoKey
	}

	if tempConfig.UseGRPC != cfg.UseGRPC {
		cfg.UseGRPC = tempConfig.UseGRPC
	}

	return nil
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

// ApplyEnvVariables - применяет переменные окружения к конфигурации.
func (cfg *Config) ApplyEnvVariables() error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}
	return nil
}

// ParseFlags - функция настройки флагов и переменных окружения сервера.
func (cfg *Config) ParseFlags() {
	flag.StringVar(&cfg.Address, "a", cfg.Address, "address and port to run agent")
	flag.StringVar(&cfg.GrpcAddress, "g", cfg.GrpcAddress, "address and port to run grpc agent")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "frequency of sending metrics to the server")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "frequency of polling metrics")
	flag.StringVar(&cfg.HashKey, "k", cfg.HashKey, "key for hash")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "Rate Limit")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path for public key file")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "Path to JSON config file")
	flag.BoolVar(&cfg.UseGRPC, "u", cfg.UseGRPC, "need to start grpc")

	flag.Parse()
}
