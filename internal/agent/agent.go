package agent

import (
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/export"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"log"
	"sync"
	"time"
)

// getMetrics -  собирает метрики и отправляет их в канал.
func getMetrics(c chan<- []models.Metrics) {
	RnMetrics := metrics.GetMetrics()
	PsMetrics, _ := metrics.GetPSMetrics()
	c <- RnMetrics
	c <- PsMetrics
}

// Run -  запускает агентов для сбора и отправки метрик.
func Run() error {
	var wg sync.WaitGroup
	cfg := envs.LoadConfig()
	err := envs.RunParameters(cfg)
	if err != nil {
		log.Println(err)
	}
	chMetrics := make(chan []models.Metrics, cfg.RateLimit)
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()
	wg.Add(1)
	go func() {
		log.Println("runtime.GetMetrics started and Ps.metrcis")
		for range pollTicker.C {
			getMetrics(chMetrics)
			log.Println("runtime.GetMetrics stoped")
		}
		wg.Done()
	}()
	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			log.Println("Report metrics started")
			for range reportTicker.C {
				export.PostQueries(cfg, workerID, chMetrics, &wg)
			}
			log.Println("Report metrics stoped")
		}()
	}
	go startTask(chMetrics)
	wg.Wait()
	return nil
}

// startTask - выполняет задачи из канала метрик.
func startTask(taskChan chan []models.Metrics) {
	for {
		select {
		case <-taskChan:
			return
		default:
			log.Println("Задача выполняется...")
		}
	}
}
