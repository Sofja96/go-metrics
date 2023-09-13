package export

import (
	"bytes"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	"github.com/Sofja96/go-metrics.git/internal/agent/metrics"
	"math/rand"
	"net/http"
	"strconv"
)

func PostQueries(cfg *envs.Config) {
	metrics.GetMetrics()
	for k, v := range metrics.ValuesGauge {
		post(cfg.Address, "gauge", k, strconv.FormatFloat(v, 'f', -1, 64))
	}
	post(cfg.Address, "counter", "PollCount", strconv.FormatUint(metrics.PollCount, 10))
	post(cfg.Address, "gauge", "RandomValue", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	metrics.PollCount = 0

}

func post(address string, t string, name string, value string) {
	bodyReader := bytes.NewReader([]byte{})

	// We can set the content type here
	fmt.Println("Running agent on", address)
	resp, err := http.Post(fmt.Sprintf("http://%s/update/%s/%s/%s", address, t, name, value), "text/plain", bodyReader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)
	fmt.Println("POST:", resp.Request)

}
