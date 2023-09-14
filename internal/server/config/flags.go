package config

import "C"
import (
	"flag"
	"os"
)

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags(s *Config) {

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		s.Address = envRunAddr
	} else {
		flag.StringVar(&s.Address, "a", "localhost:8080", "address and port to run server")
		flag.Parse()
	}
}
