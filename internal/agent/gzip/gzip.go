package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/json"

	"github.com/Sofja96/go-metrics.git/internal/models"
)

// Compress - сжимает с помощью gzip список метрик в формате JSON.
func Compress(metrics []models.Metrics) ([]byte, error) {
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
