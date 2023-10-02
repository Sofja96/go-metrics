package storage

import (
	"fmt"
)

type Gauge float64
type Counter int64

type Storage interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	GetValue(t string, name string) string
	AllMetrics() *AllMetrics
	GetCounterValue(id string) int64
	GetGaugeValue(id string) float64
	UpdateCounterData(counterData map[string]Counter)
	UpdateGaugeData(gaugeData map[string]Gauge)
}

type MemStorage struct {
	gaugeData   map[string]Gauge
	counterData map[string]Counter
}

func NewMemStorage(storeInterval int, filePath string, restore bool) *MemStorage {
	storage := MemStorage{
		gaugeData:   make(map[string]Gauge),
		counterData: make(map[string]Counter),
	}

	return &storage
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counterData[name] += Counter(value)
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.gaugeData[name] = Gauge(value)
}

func (s *MemStorage) GetValue(t string, name string) string {
	var v string
	if val, ok := s.gaugeData[name]; ok && t == "gauge" {
		v = fmt.Sprint(val)
	} else if val, ok := s.counterData[name]; ok && t == "counter" {
		v = fmt.Sprint(val)
	}
	return v
}

func (s *MemStorage) GetCounterValue(id string) int64 {
	return int64(s.counterData[id])
}

func (s *MemStorage) GetGaugeValue(id string) float64 {
	return float64(s.gaugeData[id])
}

type AllMetrics struct {
	Gauge   map[string]Gauge
	Counter map[string]Counter
}

func (s *MemStorage) AllMetrics() *AllMetrics {
	return &AllMetrics{
		Gauge:   s.gaugeData,
		Counter: s.counterData,
	}
}

func (s *MemStorage) UpdateGaugeData(gaugeData map[string]Gauge) {
	s.gaugeData = gaugeData
}

func (s *MemStorage) UpdateCounterData(counterData map[string]Counter) {
	s.counterData = counterData
}
