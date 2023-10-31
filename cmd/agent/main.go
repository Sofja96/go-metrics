package main

import (
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/export"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"log"
	"time"
)

func main() {
	cfg := envs.LoadConfig()
	err := envs.RunParameters(cfg)
	if err != nil {
		log.Fatal(err)
	}
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			metrics.GetMetrics()
			b, _ := json.Marshal(metrics.ValuesGauge)
			fmt.Println(string(b))
		case <-reportTicker.C:
			export.PostQueries(cfg)
			export.PostQueriesMetrics(cfg)
		}
	}
}
