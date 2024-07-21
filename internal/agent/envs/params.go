package envs

import (
	"flag"
	"os"
	"strconv"
)

// RunParameters - функция настройки флагов и переменных окружения агента.
func RunParameters(cfg *Config) error {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&cfg.PollInterval, "p", 2, "frequency of polling metrics")
	flag.StringVar(&cfg.HashKey, "k", "", "key for hash")
	flag.IntVar(&cfg.RateLimit, "l", 2, "Rate Limit")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.Address = envRunAddr
	}
	if envRunAddr := os.Getenv("REPORT_INTERVAL"); envRunAddr != "" {
		cfg.ReportInterval, _ = strconv.Atoi(envRunAddr)
	}
	if envRunAddr := os.Getenv("POLL_INTERVAL"); envRunAddr != "" {
		cfg.PollInterval, _ = strconv.Atoi(envRunAddr)
	}
	if envRunAddr := os.Getenv("KEY"); envRunAddr != "" {
		cfg.HashKey = envRunAddr
	}
	if envRunAddr := os.Getenv("RATE_LIMIT"); envRunAddr != "" {
		cfg.RateLimit, _ = strconv.Atoi(envRunAddr)
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 1
	}

	return nil
}
