package export

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/hash"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/models"
)

// Настройки повторной отправки по умолчанию.
const (
	retryMax     int           = 3               // максимальное количество
	retryWaitMin time.Duration = time.Second * 1 // минимальное время ожидания
	retryWaitMax time.Duration = time.Second * 5 // максимальное время ожидания
)

// PostQueries - функция для формирования метрик перед отправкой и запуска отправки метрик.
func PostQueries(cfg *envs.Config, workerID int, chIn <-chan []models.Metrics, wg *sync.WaitGroup) {
	metrics.GetMetrics()
	allMetrics := make([]models.Metrics, 0, len(metrics.GetMetrics()))
	log.Println("Running agent on", cfg.Address)
	log.Println("workerID", strconv.Itoa(workerID), "SendMetricWorker started")
	url := fmt.Sprintf("http://%s/updates/", cfg.Address)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Backoff = linearBackoff

	metrics.Mu.Lock()
	defer metrics.Mu.Unlock()

	for k, v := range metrics.ValuesGauge {
		val := v // создаем локальную переменную value
		allMetrics = append(allMetrics, models.Metrics{
			MType: "gauge",
			ID:    k,
			Value: &val, // передаем указатель на локальную переменную value
		})
		log.Printf("%s: %d", k, int(val))
	}
	for k, v := range metrics.ValuesCounter {
		val := v
		allMetrics = append(allMetrics, models.Metrics{MType: "counter", ID: k, Delta: &val})
		log.Printf("%s: %d", k, int(val))
	}
	gz, _ := compress(allMetrics)
	postBatch(retryClient, url, cfg.HashKey, gz)
	wg.Done()
}

// postBatch - функция отправки сжатых метрик на сервер.
func postBatch(r *retryablehttp.Client, url string, key string, m []byte) error {
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

// compress - сжимает с помощью gzip список метрик в формате JSON.
func compress(metrics []models.Metrics) ([]byte, error) {
	var b bytes.Buffer
	js, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	_, err = gz.Write(js)
	if err != nil {
		return nil, err
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// linearBackoff - расчитывает время ожижания между попытками отправки
func linearBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleepTime := min + min*time.Duration(2*attemptNum)
	return sleepTime
}
