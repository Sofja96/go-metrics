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
)

func PostQueriesMetrics(cfg *envs.Config) {
	metrics.GetMetrics()
	url := fmt.Sprintf("http://%s/update/", cfg.Address)
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Backoff = linearBackoff
	for k, v := range metrics.ValuesGauge {
		post(retryClient, url, cfg.HashKey, models.Metrics{MType: "gauge", ID: k, Value: &v})
	}
	for k, v := range metrics.ValuesCounter {
		post(retryClient, url, cfg.HashKey, models.Metrics{MType: "counter", ID: k, Delta: &v})
	}
}

func post(r *retryablehttp.Client, url string, key string, m models.Metrics) error {
	gz, err := compressMetrics(m)
	if err != nil {
		return fmt.Errorf("error on compressing metrics on request: %w", err)
	}
	req, err := retryablehttp.NewRequest("POST", url, gz)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")
	if len(key) != 0 {
		req.Header.Set("HashSHA256", hash.ComputeHmac256([]byte(key), gz))
		log.Println(hash.ComputeHmac256([]byte(key), gz), "compute")
		log.Println([]byte(key), key)
		log.Println(gz, "gz")
		log.Println(m, "m")
	}
	resp, err := r.Do(req)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}
	defer resp.Body.Close()
	log.Println(resp.StatusCode)
	log.Println(resp.Header)
	log.Println(resp.Body)

	return nil
}

func compressMetrics(metrics models.Metrics) ([]byte, error) {
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
