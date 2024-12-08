package memory

import (
	"errors"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	"log"
	"os"
	"sync"
)

type Gauge float64
type Counter int64

type MemStorage struct {
	gaugeData   map[string]Gauge
	counterData map[string]Counter
	mutex       sync.RWMutex
}

func (s *MemStorage) Ping() error {
	return nil
}

// NewInMemStorage - создает локальное хранилище
func NewInMemStorage(storeInterval int, filePath string, restore bool) (storage.Storage, error) {
	return NewMemStorage(storeInterval, filePath, restore)
}

func NewMemStorage(storeInterval int, filePath string, restore bool) (*MemStorage, error) {
	s := &MemStorage{
		gaugeData:   make(map[string]Gauge),
		counterData: make(map[string]Counter),
	}

	if restore {
		err := LoadStorageFromFile(s, filePath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("failed to restore data from file: %w", err)
			}
		}
	}

	if storeInterval != 0 {
		go func() {
			err := Dump(s, filePath, storeInterval)
			if err != nil {
				log.Print(err)
			}
		}()
	}
	return s, nil

}

func (s *MemStorage) UpdateCounter(name string, value int64) (int64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.counterData[name] += Counter(value)
	return int64(s.counterData[name]), nil
}

func (s *MemStorage) UpdateGauge(name string, value float64) (float64, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.gaugeData[name] = Gauge(value)
	return float64(s.gaugeData[name]), nil
}

func (s *MemStorage) GetCounterValue(id string) (int64, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	val, ok := s.counterData[id]
	return int64(val), ok
}

func (s *MemStorage) GetGaugeValue(id string) (float64, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok := s.gaugeData[id]
	return float64(s.gaugeData[id]), ok
}

type AllMetrics struct {
	Gauge   map[string]Gauge
	Counter map[string]Counter
}

func (s *MemStorage) AllMetrics() *AllMetrics {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return &AllMetrics{
		Gauge:   s.gaugeData,
		Counter: s.counterData,
	}
}

func (s *MemStorage) UpdateGaugeData(gaugeData map[string]Gauge) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.gaugeData = gaugeData
}

func (s *MemStorage) UpdateCounterData(counterData map[string]Counter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.counterData = counterData
}

func (s *MemStorage) GetAllGauges() ([]storage.GaugeMetric, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	gauges := make([]storage.GaugeMetric, 0, len(s.gaugeData))
	for name, value := range s.gaugeData {
		gauges = append(gauges, storage.GaugeMetric{Name: name, Value: float64(value)})
	}
	return gauges, nil
}

// GetAllCounters returns all counter metrics.
func (s *MemStorage) GetAllCounters() ([]storage.CounterMetric, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	counters := make([]storage.CounterMetric, 0, len(s.counterData))
	for name, value := range s.counterData {
		counters = append(counters, storage.CounterMetric{Name: name, Value: int64(value)})
	}

	return counters, nil
}

func (s *MemStorage) BatchUpdate(metrics []models.Metrics) error {
	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			_, err := s.UpdateGauge(v.ID, *v.Value)
			if err != nil {
				return fmt.Errorf("error update gauge for batch update: %v", err)
			}
		case "counter":
			val, err := s.UpdateCounter(v.ID, *v.Delta)
			if err != nil {
				return fmt.Errorf("error update counter for batch update: %v", err)
			}
			*v.Delta = val
		default:
			return fmt.Errorf("unsupported metrics type: %s", v.MType)

		}
	}
	return nil
}
