package envs

// Config - струтура хранения настроек агента.
type Config struct {
	Address        string `env:"ADDRESS"`         // адрес работы агента сбора метрик
	ReportInterval int    `env:"REPORT_INTERVAL"` // интервал отправки метрик
	PollInterval   int    `env:"POLL_INTERVAL"`   // интервал сбора метрик
	HashKey        string `env:"KEY"`             // ключ аутентификации
	RateLimit      int    `env:"RATE_LIMIT"`      // ограничение на количество исходящих запросов
	CryptoKey      string `env:"CRYPTO_KEY"`      // файл с публичным ключом сервера
}

// LoadConfig - загружает конфигурацию агента.
func LoadConfig() *Config {
	var cfg Config
	return &cfg

}
