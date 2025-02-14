package storage

import (
	"context"

	"github.com/Sofja96/go-metrics.git/internal/models"
)

// Storage - интерфейс хранилища
type Storage interface {
	// UpdateCounter - обновляет метрику типа counter
	UpdateCounter(ctx context.Context, name string, value int64) (int64, error)
	// UpdateGauge - обновляет метрику типа gauge
	UpdateGauge(ctx context.Context, name string, value float64) (float64, error)
	// GetCounterValue - получает метрику типа counter
	GetCounterValue(ctx context.Context, id string) (int64, bool)
	// GetGaugeValue - получает метрику типа gauge
	GetGaugeValue(ctx context.Context, id string) (float64, bool)
	// Ping - проверяет доступность хринилища
	Ping(context.Context) error
	// GetAllGauges - получает все метрики типа gauges
	GetAllGauges(context.Context) ([]GaugeMetric, error)
	// GetAllCounters - получает все метрики типа counter
	GetAllCounters(context.Context) ([]CounterMetric, error)
	// BatchUpdate - обновляет метрики пачкой
	BatchUpdate(ctx context.Context, metrics []models.Metrics) error
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
