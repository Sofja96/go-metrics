package agent

import (
	"log"
	"sync"
	"time"

	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/export"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
)

// getMetrics -  собирает метрики и отправляет их в канал.
func getMetrics(collector *metrics.Metrics, c chan<- []byte) {
	err := collector.GetMetrics()
	if err != nil {
		log.Printf("Error collecting runtime metrics: %v", err)
	}

	err = collector.GetPSMetrics()
	if err != nil {
		log.Printf("Error collecting PS metrics: %v", err)
	}

	compressedMetrics, err := collector.PrepareMetrics()
	if err != nil {
		log.Printf("Error preparing metrics: %v", err)
		return
	}

	c <- compressedMetrics
}

// Run -  запускает агентов для сбора и отправки метрик.
func Run() error {
	var wg sync.WaitGroup
	collector := metrics.NewMetricsCollector()
	cfg := envs.LoadConfig()
	err := envs.RunParameters(cfg)
	if err != nil {
		log.Println(err)
	}

	chMetrics := make(chan []byte, cfg.RateLimit)

	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("runtime.GetMetrics started and Ps.metrics")
		for range pollTicker.C {
			getMetrics(collector, chMetrics)
		}
		log.Println("runtime.GetMetrics stopped")
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		startTask(chMetrics)
	}()

	wg.Wait()
	close(chMetrics)
	return nil
}

// startTask - выполняет задачи из канала метрик.
func startTask(taskChan chan []byte) {
	for {
		select {
		case <-taskChan:
			log.Println("задача завершена")
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
