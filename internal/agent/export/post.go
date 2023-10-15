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
)

func PostQueries(cfg *envs.Config) {
	metrics.GetMetrics()
	allMetrics := make([]models.Metrics, 0, len(metrics.GetMetrics()))
	log.Println("Running agent on", cfg.Address)
	url := fmt.Sprintf("http://%s/updates/", cfg.Address)
	ro := grequests.RequestOptions{
		Headers: map[string]string{
			"content-type":     "application/json",
			"content-encoding": "gzip",
			"Accept-Encoding":  "gzip",
		},
	}
	session := grequests.NewSession(&ro)
	for k, v := range metrics.ValuesGauge {
		allMetrics = append(allMetrics, models.Metrics{MType: "gauge", ID: k, Value: &v})
		//	post(session, url, allMetrics)
	}

	for k, v := range metrics.ValuesCounter {
		allMetrics = append(allMetrics, models.Metrics{MType: "counter", ID: k, Delta: &v})
	}
	//var pc = int64(metrics.PollCount)
	//allMetrics = append(allMetrics, models.Metrics{MType: "counter", ID: "PollCount", Delta: &pc})
	//post(session, url, allMetrics)
	//r := rand.Float64()
	//allMetrics = append(allMetrics, models.Metrics{MType: "gauge", ID: "RandomValue", Value: &r})
	post(session, url, allMetrics)
	//metrics.PollCount = 0

}

func post(s *grequests.Session, url string, m []models.Metrics) {
	gz, err := compress(m)
	if err != nil {
		log.Printf("error on compressing metrics on request: %v", err)
	}
	resp, err := s.Post(url, &grequests.RequestOptions{JSON: gz})
	if err != nil {
		log.Println(err)
	}
	log.Println(resp.StatusCode)
	log.Println(resp.Header)
	log.Println(resp.RawResponse)
}

func compress(metrics []models.Metrics) ([]byte, error) {
	var b bytes.Buffer
	js, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("impossible to marshall metrics: %s", err)
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
