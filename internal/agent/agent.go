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
func getMetrics(c chan<- []byte, mutex *sync.Mutex) {
	//defer close(c)
	m := metrics.NewMetricsCollector()
	mutex.Lock()
	defer mutex.Unlock()
	_, err := m.GetMetrics()
	if err != nil {
		log.Println("Error collecting runtime metrics:", err)
	}
	_, err = m.GetPSMetrics()
	if err != nil {
		log.Println("Error collecting system metrics:", err)
	}

	data, err := metrics.PrepareMetrics(m)
	if err != nil {
		log.Println("Error preparing metrics:", err)
		//return err
	}
	c <- data
	//close(c)
}

// getMetrics -  собирает метрики и отправляет их в канал.
//func getMetrics(mutex *sync.Mutex) chan<- []byte {
//	//defer close(c)
//	input := make(chan []byte)
//	m := metrics.NewMetricsCollector()
//	mutex.Lock()
//	defer mutex.Unlock()
//	RnMetrics, _ := m.GetMetrics()
//	go func() {
//		input <- RnMetrics
//		close(input)
//	}()
//	//err := m.GetMetrics()
//	//if err != nil {
//	//	log.Println("Error collecting runtime metrics:", err)
//	//}
//	//err = m.GetPSMetrics()
//	//if err != nil {
//	//	log.Println("Error collecting system metrics:", err)
//	//}
//	//
//	//data, err := m.PrepareMetrics()
//	//if err != nil {
//	//	log.Println("Error preparing metrics:", err)
//	//	return
//	//}
//	//c <- data
//	//close(c)
//}

// Run -  запускает агентов для сбора и отправки метрик.
func Run() error {
	var wg sync.WaitGroup
	var m sync.Mutex
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
		//defer close(chMetrics)
		log.Println("runtime.GetMetrics started and Ps.metrcis")
		for range pollTicker.C {
			getMetrics(chMetrics, &m)
			//log.Println("runtime.GetMetrics stoped")
		}
		close(chMetrics)
		log.Println("runtime.GetMetrics stoped")
	}()

	//close(chMetrics)

	for i := 0; i < cfg.RateLimit; i++ {
		log.Println("Rate limit: ", cfg.RateLimit)
		wg.Add(1)
		workerID := i
		go func() {
			//defer wg.Done()
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
	//defer close(chMetrics)
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

//func merge(outputChan ...<-chan []byte) <-chan []byte {
//	var wg sync.WaitGroup
//	chOut := make(chan []byte, len(outputChan))
//
//	wg.Add(len(outputChan))
//
//	output := func(ch <-chan []byte) {
//		for v := range ch {
//			chOut <- v
//		}
//		wg.Done()
//	}
//
//	for _, ch := range outputChan {
//		go output(ch)
//	}
//
//	go func() {
//		wg.Wait()
//		close(chOut)
//	}()
//
//	return chOut
//}
