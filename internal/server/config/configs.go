package config

// Config - структура хранения настроек сервера.
type Config struct {
	Address       string `env:"ADDRESS"`           // адрес и порт работы сервера
	StoreInterval int    `env:"STORE_INTERVAL"`    // интервал сохранения метрик
	FilePath      string `env:"FILE_STORAGE_PATH"` // путь к файлу хранилища
	Restore       bool   `env:"RESTORE"`           // указывает необходимость восстановить данные при старте сервера
	DatabaseDSN   string `env:"DATABASE_DSN"`      // строка подключения к БД
	HashKey       string `env:"KEY"`               // ключ аутентификации
}

// LoadConfig - загружает конфигурацию сервера.
func LoadConfig() *Config {
	var cfg Config
	return &cfg

}
