package config

import "C"
import (
	"flag"
	"os"
	"strconv"
)

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags(s *Config) {
	//var conf Config
	flag.StringVar(&s.Address, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&s.FilePath, "f", "/tmp/metris-db.json", "path file storage to save data")
	flag.IntVar(&s.StoreInterval, "i", 300, "interval for saving metrics on the server")
	flag.BoolVar(&s.Restore, "r", true, "need to load data at startup")
	flag.StringVar(&s.DatabaseDSN, "d", "", "connect to database")
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
}
