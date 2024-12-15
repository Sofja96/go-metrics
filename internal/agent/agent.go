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
func getMetrics(c chan<- []byte) {
	m := metrics.NewMetricsCollector()
	err := m.GetMetrics()
	if err != nil {
		log.Println("Error collecting runtime metrics:", err)
	}
	err = m.GetPSMetrics()
	if err != nil {
		log.Println("Error collecting system metrics:", err)
	}

	data, err := m.PrepareMetrics()
	if err != nil {
		log.Println("Error preparing metrics:", err)
		return
	}

	c <- data
}

// Run -  запускает агентов для сбора и отправки метрик.
func Run() error {
	var wg sync.WaitGroup
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
		log.Println("runtime.GetMetrics started and Ps.metrcis")
		for range pollTicker.C {
			getMetrics(chMetrics)
			log.Println("runtime.GetMetrics stoped")
		}
	}()
	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		workerID := i
		go func() {
			defer wg.Done()
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
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
