package storage

import (
	"github.com/Sofja96/go-metrics.git/internal/models"
)

//import "github.com/Sofja96/go-metrics.git/internal/storage/memory"

type Storage interface {
	UpdateCounter(name string, value int64) (int64, error)
	UpdateGauge(name string, value float64) (float64, error)
	//AllMetrics() *memory.AllMetrics
	GetCounterValue(id string) (int64, bool)
	GetGaugeValue(id string) (float64, bool)
	Ping() error
	GetAllGauges() ([]GaugeMetric, error)
	GetAllCounters() ([]CounterMetric, error)
	//UpdateGauges(metrics []GaugeMetric) ([]GaugeMetric, error)
	//UpdateCounters(metrics []CounterMetric) ([]CounterMetric, error)
	BatchUpdate(metrics []models.Metrics) error
}

type CounterMetric struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type GaugeMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Metrics struct {
	Gauges   []GaugeMetric
	Counters []CounterMetric
}
