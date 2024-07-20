package envs

type Config struct {
	Address        string `env:"ADDRESS"`         // адрес работы агента сбора метрик
	ReportInterval int    `env:"REPORT_INTERVAL"` // интервал отправки метрик
	PollInterval   int    `env:"POLL_INTERVAL"`   // интервал сбора метрик
	HashKey        string `env:"KEY"`             // ключ аутентификации
	RateLimit      int    `env:"RATE_LIMIT"`      // ограничение на количество исходящих запросов
}

func LoadConfig() *Config {
	var cfg Config
	return &cfg

}
