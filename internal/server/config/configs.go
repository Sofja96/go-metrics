package config

type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	FilePath      string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	DatabaseDSN   string `env:"DATABASE_DSN"`
	HashKey       string `env:"KEY"`
}

func LoadConfig() *Config {
	var cfg Config
	return &cfg

}
