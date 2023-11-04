package main

import (
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/export"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"log"
	"sync"
	"time"
)

//func getMetrics(c chan<- []models.Metrics) {
//	//chOut := make(chan []models.Metrics, 100)
//	m := metrics.GetMetrics()
//	c <- m
//}
//
//func getMetricsPs(c chan<- []models.Metrics) {
//	m := metrics.GetPSMetrics()
//	c <- m
//}

func main() {
	var wg sync.WaitGroup
	cfg := envs.LoadConfig()
	err := envs.RunParameters(cfg)
	if err != nil {
		log.Fatal(err)
	}
	c := make(chan []models.Metrics, 100)
	chOut := make(chan error, cfg.RateLimit)
	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()
	wg.Add(1)
	go func() {
		log.Println("runtime.GetMetrics started")

		for range pollTicker.C {
			c <- metrics.GetMetrics()
			b, _ := json.Marshal(metrics.ValuesGauge)
			fmt.Println(string(b))
			msg := <-c
			log.Println(msg, "runtime.GetMetrics")
			//getMetrics(c)
			log.Println("runtime.GetMetrics stoped")
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		log.Println("Ps.GetMetrics started")
		for range pollTicker.C {
			c <- metrics.GetPSMetrics()
			b, _ := json.Marshal(metrics.ValuesGauge)
			fmt.Println(string(b))
			msg := <-c
			log.Println(msg, "Ps.GetMetrics")
			log.Println("Ps.GetMetrics stoped")
		}
		wg.Done()
	}()
	for i := 0; i < cfg.RateLimit; i++ {
		workerID := i
		go func() {
			log.Println("Report metrics started")
			for range reportTicker.C {
				msg := <-c
				log.Println(msg, "channel")
				export.PostQueries(cfg, workerID, c, chOut)
			}
			log.Println("Report metrics stoped")
		}()
	}
	wg.Wait()
}

//for {
//	select {
//	case <-pollTicker.C:
//		metrics.GetMetrics()
//		b, _ := json.Marshal(metrics.ValuesGauge)
//		fmt.Println(string(b))
//	case <-reportTicker.C:
//		export.PostQueries(cfg)
//	}
//}
