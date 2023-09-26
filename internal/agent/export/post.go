package export

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/levigross/grequests"
	"log"
	"math/rand"
)

func PostQueries(cfg *envs.Config) {
	metrics.GetMetrics()
	fmt.Println("Running agent on", cfg.Address)
	url := fmt.Sprintf("http://%s/update/", cfg.Address)
	ro := grequests.RequestOptions{
		Headers: map[string]string{
			"content-type":     "application/json",
			"content-encoding": "gzip",
			//"Accept-Encoding":  "gzip",
		},
	}
	session := grequests.NewSession(&ro)
	for k, v := range metrics.ValuesGauge {
		post(session, url, models.Metrics{MType: "gauge", ID: k, Value: &v})
	}
	var pc = int64(metrics.PollCount)
	post(session, url, models.Metrics{MType: "counter", ID: "PollCount", Delta: &pc})
	r := rand.Float64()
	post(session, url, models.Metrics{MType: "gauge", ID: "RandomValue", Value: &r})
	metrics.PollCount = 0

}

func post(s *grequests.Session, url string, m models.Metrics) {
	gz, err := compress(m)
	if err != nil {
		log.Fatalf("error on compressing metrics on request: %v", err)
	}
	resp, err := s.Post(url, &grequests.RequestOptions{JSON: gz})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header)
	fmt.Println(resp.RawResponse)
}

func compress(metrics models.Metrics) ([]byte, error) {
	var b bytes.Buffer
	js, err := json.Marshal(metrics)
	if err != nil {
		log.Fatalf("impossible to marshall metrics: %s", err)
	}
	gz, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	_, err = gz.Write(js)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = gz.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}
