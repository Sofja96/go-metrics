package export

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Sofja96/go-metrics.git/internal/agent/envs"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/Sofja96/go-metrics.git/internal/agent/hash"
)

// Настройки повторной отправки по умолчанию.
const (
	retryMax     int           = 3               // максимальное количество
	retryWaitMin time.Duration = time.Second * 1 // минимальное время ожидания
	retryWaitMax time.Duration = time.Second * 5 // максимальное время ожидания
)

// PostQueries - функция для формирования метрик перед отправкой и запуска отправки метрик.
func PostQueries(cfg *envs.Config, workerID int, chIn <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("workerID", workerID, "started")

	url := fmt.Sprintf("http://%s/updates/", cfg.Address)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Backoff = linearBackoff

	for compressedData := range chIn {
		err := PostBatch(retryClient, url, cfg.HashKey, compressedData)
		if err != nil {
			log.Printf("Error posting metrics: %v", err)
		}
	}
}

// PostBatch - функция отправки сжатых метрик на сервер.
func PostBatch(r *retryablehttp.Client, url string, key string, m []byte) error {
	req, err := retryablehttp.NewRequest("POST", url, m)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")

	if len(key) != 0 {
		hmac, err := hash.ComputeHmac256([]byte(key), m)
		if err != nil {
			return fmt.Errorf("error compute hash data: %w", err)
		}
		req.Header.Add("HashSHA256", hmac)
	}
	resp, err := r.Do(req)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status Code: %d\n", resp.StatusCode)
	log.Printf("Response Headers: %v\n", resp.Header)

	return nil
}

// linearBackoff - расчитывает время ожижания между попытками отправки
func linearBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleepTime := min + min*time.Duration(2*attemptNum)
	return sleepTime
}
