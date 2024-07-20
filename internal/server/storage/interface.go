package storage

import (
	"github.com/Sofja96/go-metrics.git/internal/models"
	"io"
)

// Storage - интерфейс хранилища
type Storage interface {
	// UpdateCounter - обновляет метрику типа counter
	UpdateCounter(name string, value int64) (int64, error)
	// UpdateGauge - обновляет метрику типа gauge
	UpdateGauge(name string, value float64) (float64, error)
	// GetCounterValue - получает метрику типа counter
	GetCounterValue(id string) (int64, bool)
	// GetGaugeValue - получает метрику типа gauge
	GetGaugeValue(id string) (float64, bool)
	// Ping - проверяет доступность хринилища
	Ping() error
	// GetAllGauges - получает все метрики типа gauges
	GetAllGauges() ([]GaugeMetric, error)
	// GetAllCounters - получает все метрики типа counter
	GetAllCounters() ([]CounterMetric, error)
	// BatchUpdate - обновляет метрики пачкой
	BatchUpdate(w io.Writer, metrics []models.Metrics) error
}

// CounterMetric - структура метрик counter, содержащая имя и значение
type CounterMetric struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// GaugeMetric  - структура метрик gauge, содержащая имя и значение
type GaugeMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}
