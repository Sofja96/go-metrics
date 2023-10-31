package export

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/hash"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/hashicorp/go-retryablehttp"
	"log"
	"net/http"
	"time"
)

const (
	retryMax     int           = 3
	retryWaitMin time.Duration = time.Second * 1
	retryWaitMax time.Duration = time.Second * 5
)

func PostQueries(cfg *envs.Config) {
	metrics.GetMetrics()
	allMetrics := make([]models.Metrics, 0, len(metrics.GetMetrics()))
	log.Println("Running agent on", cfg.Address)
	url := fmt.Sprintf("http://%s/updates/", cfg.Address)
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Backoff = linearBackoff
	for k, v := range metrics.ValuesGauge {
		allMetrics = append(allMetrics, models.Metrics{MType: "gauge", ID: k, Value: &v})
	}
	for k, v := range metrics.ValuesCounter {
		allMetrics = append(allMetrics, models.Metrics{MType: "counter", ID: k, Delta: &v})
	}
	postBatch(retryClient, url, cfg.HashKey, allMetrics)
	log.Println(cfg.HashKey)
}

func postBatch(r *retryablehttp.Client, url string, key string, m []models.Metrics) error {
	gz, err := compress(m)
	if err != nil {
		return fmt.Errorf("error on compressing metrics on request: %w", err)
	}
	//hashedMetrics, err := hash.ComputeHmac256([]byte(key), gz)
	//if err != nil {
	//	return fmt.Errorf("error calculating hmac: %w", err)
	//}
	fmt.Println(hash.ComputeHmac256([]byte(key), gz))
	req, err := retryablehttp.NewRequest("POST", url, bytes.NewReader(gz))
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")
	//req.Header.Add("HashSHA256", hash.ComputeHmac256([]byte(key), gz))
	//req.Header.Set("HashSHA256", hash.ComputeHmac256([]byte(key), gz))
	//req.Header.Set("HashSHA256", hashedMetrics)
	if len(key) != 0 {
		req.Header.Set("HashSHA256", hash.ComputeHmac256([]byte(key), gz))
		log.Println(hash.ComputeHmac256([]byte(key), gz), "compute")
		log.Println([]byte(key), key)
		log.Println(gz, "gz")
		log.Println(m, "m")
	}
	//req.SetBody(hashedMetrics)
	//req.SetBody(gz)
	//req = req.
	fmt.Println(req.Header)
	//req.Header.Set("HashSHA256", hash.ComputeHmac256([]byte(key), gz))
	resp, err := r.Do(req)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	log.Println(resp.Header)
	log.Println(resp.Body)
	log.Println(resp.ContentLength, "resp")
	log.Println(req.ContentLength, "req")
	log.Println()

	return nil
}

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

func linearBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleepTime := min + min*time.Duration(2*attemptNum)
	return sleepTime
}
