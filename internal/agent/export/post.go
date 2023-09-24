package export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"log"
	"math/rand"
	"net/http"
)

func PostQueries(cfg *envs.Config) {
	metrics.GetMetrics()
	url := "http://%s/update/"
	for k, v := range metrics.ValuesGauge {
		post(url, cfg.Address, models.Metrics{MType: "gauge", ID: k, Value: &v})
	}
	var pc = int64(metrics.PollCount)
	post(url, cfg.Address, models.Metrics{MType: "counter", ID: "PollCount", Delta: &pc})
	r := rand.Float64()
	post(url, cfg.Address, models.Metrics{MType: "gauge", ID: "RandomValue", Value: &r})
	metrics.PollCount = 0

}

func post(url, address string, m models.Metrics) {
	marshalled, err := json.Marshal(m)
	if err != nil {
		log.Fatalf("impossible to marshall teacher: %s", err)
	}
	//http://%s/update/%s/%s/%s

	// We can set the content type here
	fmt.Println("Running agent on", address)
	resp, err := http.Post(fmt.Sprintf(url, address), "application/json", bytes.NewReader(marshalled))
	//
	////resp, err := http.Post(fmt.Sprintf("http://%s/update/%s/%s/%s", address, t, name, value), "application/json", bytes.NewReader(marshalled))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("POST:", resp.Request)
	fmt.Println("Body: ", resp.Body)

}
