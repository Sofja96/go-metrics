package envs

import (
	"flag"
	"os"
	"strconv"
)

func RunParameters(cfg *Config) error {
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&cfg.PollInterval, "p", 2, "frequency of polling metrics")
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

	return nil
}
