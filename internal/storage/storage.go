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
}

type MemStorage struct {
	gaugeData   map[string]Gauge
	counterData map[string]Counter
}

// NewMemStorage returns a new in memory storage instance.
func NewMemStorage() *MemStorage {

	return &MemStorage{
		gaugeData:   make(map[string]Gauge),
		counterData: make(map[string]Counter),
	}
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
	Gauge   *map[string]Gauge
	Counter *map[string]Counter
}

func (s *MemStorage) AllMetrics() *AllMetrics {
	return &AllMetrics{
		Gauge:   &s.gaugeData,
		Counter: &s.counterData,
	}
}

//TODO

// Мапы всегда передаются по ссылке, то есть за пределами MemStorage, например, в хендлерах,
// мы сможем как-то изменять их. Для сдачи работы не критично,
// но я бы еще добавил копирование мап и возвращал новые, или возвращал бы собственные структуры
//type GaugeMetric struct {
//	Name  string
//	Value float64
//}
//type CounterMetric struct {
//	Name  string
//	Value int64
//}
//type AllMetrics struct {
//	Gauges   map[string]GaugeMetric
//	Counters map[string]CounterMetric
//}

//func (s *MemStorage) AllMetrics(name string, value float64) *AllMetrics {
//	return &AllMetrics{
//		// s.gaugeData: &Gauges
//		Gauges:   make(map[string]GaugeMetric),
//		 Gauges.Name = s.gaugeData[name]
//
//
//
//
//		//Counters: make(map[string]CounterMetric),
//		//	Gauges: s.gaugeData,
//		//Counters: &[]CounterMetric,
//	}
//}
