package config

import (
	"flag"
	"os"
	"strconv"
)

// ParseFlags - функция настройки флагов и переменных окружения сервера.
func ParseFlags(s *Config) {
	flag.StringVar(&s.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&s.FilePath, "f", "/tmp/metrics-db.json", "path file storage to save data")
	flag.IntVar(&s.StoreInterval, "i", 3, "interval for saving metrics on the server")
	flag.BoolVar(&s.Restore, "r", true, "need to load data at startup")
	flag.StringVar(&s.DatabaseDSN, "d", "", "connect to database")
	flag.StringVar(&s.HashKey, "k", "", "key for hash")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.Address = envRunAddr
	}
	if envRunAddr := os.Getenv("FILE_STORAGE_PATH"); envRunAddr != "" {
		s.FilePath = envRunAddr
	}
	if envRunAddr := os.Getenv("STORE_INTERVAL"); envRunAddr != "" {
		s.StoreInterval, _ = strconv.Atoi(envRunAddr)
	}
	if envRunAddr := os.Getenv("RESTORE"); envRunAddr != "" {
		s.Restore, _ = strconv.ParseBool(envRunAddr)
	}
	if envRunAddr := os.Getenv("DATABASE_DSN"); envRunAddr != "" {
		s.DatabaseDSN = envRunAddr
	}
	if envRunAddr := os.Getenv("KEY"); envRunAddr != "" {
		s.HashKey = envRunAddr
	}
}
