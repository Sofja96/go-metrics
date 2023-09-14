package storage

import (
	"fmt"
)

type gauge float64
type Counter int64

type Metrics interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	GetValue(t string, name string) string
	AllMetrics() *AllMetrics
}

type MemStorage struct {
	gaugeData   map[string]gauge
	counterData map[string]Counter
}

// NewMemStorage returns a new in memory storage instance.
func NewMemStorage() *MemStorage {

	return &MemStorage{
		gaugeData:   make(map[string]gauge),
		counterData: make(map[string]Counter),
	}
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.counterData[name] += Counter(value)
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.gaugeData[name] = gauge(value)
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

type AllMetrics struct {
	Gauge   map[string]gauge
	Counter map[string]Counter
}

func (s *MemStorage) AllMetrics() *AllMetrics {
	return &AllMetrics{
		Gauge:   s.gaugeData,
		Counter: s.counterData,
	}
}
