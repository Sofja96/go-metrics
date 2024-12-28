package agent

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	cfg, err := envs.LoadConfig()
	if err != nil {
		log.Printf("error load config: %v", err)
	}

	publicKey, err := LoadPublicKey(cfg.CryptoKey)
	if err != nil {
		return fmt.Errorf("failed to load public key: %w", err)
	}

	chMetrics := make(chan []byte, cfg.RateLimit)

	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sig := <-signalChan
		log.Printf("Получен сигнал: %s. Завершаем работу...", sig)
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Сбор метрик.")
		for {
			select {
			case <-ctx.Done():
				log.Println("Сбор метрик завершен.")
				return
			case <-pollTicker.C:
				getMetrics(collector, chMetrics)
			}
		}
	}()
	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		workerID := i
		go func(workerId int) {
			defer wg.Done()
			log.Println("Отправка метрик.")
			for {
				select {
				case <-ctx.Done():
					log.Println("Отправка метрик остановлена по отмене контекста.")
					return
				case <-reportTicker.C:
					export.PostQueries(ctx, cfg, workerID, chMetrics, publicKey)
				}
			}
		}(workerID)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTask(ctx, chMetrics)
	}()

	wg.Wait()

	close(chMetrics)
	log.Println("Все метрики успешно отправлены.")
	return nil
}

// startTask - выполняет задачи из канала метрик.
func startTask(ctx context.Context, taskChan chan []byte) {
	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				return
			}
			log.Printf("Задача получена: %v", task)

		case <-ctx.Done():
			log.Println("Контекст отменен, завершение работы.")
			return

		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// LoadPublicKey - функция загрузки публичного ключа из файла.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading public key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM format or missing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %w", err)
	}

	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaKey, nil
}
