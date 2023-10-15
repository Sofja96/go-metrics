package memory

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/Sofja96/go-metrics.git/internal/storage/database"
	"io"
	"log"
	"os"
)

type Gauge float64
type Counter int64

//type Storage interface {
//	UpdateCounter(name string, value int64)
//	UpdateGauge(name string, value float64)
//	AllMetrics() *AllMetrics
//	GetCounterValue(id string) (int64, bool)
//	GetGaugeValue(id string) (float64, bool)
//}

type MemStorage struct {
	gaugeData   map[string]Gauge
	counterData map[string]Counter
	//FileStorage *FileStorage
}

func (s *MemStorage) Ping() error {
	return nil
}

func NewInMemStorage(storeInterval int, filePath string, restore bool) (storage.Storage, error) {
	return NewMemStorage(storeInterval, filePath, restore)
}

func NewPostgresqlStorage(dsn string) (storage.Storage, error) {
	return database.NewStorage(dsn)
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
		//if err != nil {
		//	return nil, err
		//}
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

	//	return s
}

//func NewMemFileStorage (storeInterval int, filePath string, restore bool) (storage.Storage, error) {
//	var metrics FileStorage.
//}

func (s *MemStorage) UpdateCounter(name string, value int64) (int64, error) {
	s.counterData[name] += Counter(value)
	return value, nil
}

//func (s *MemStorage) UpdateCounters(metrics []storage.CounterMetric) ([]storage.CounterMetric, error) {
//	for _, v := range metrics {
//		s.counterData[v.Name] += Counter(v.Value)
//		//switch v.MType {
//		//case "gauge":
//		//	s.UpdateGauge(v.ID, *v.Value)
//		//case "counter":
//		//	s.UpdateCounter(v.ID, *v.Delta)
//
//	}
//	return metrics, nil
//}

func (s *MemStorage) UpdateGauge(name string, value float64) (float64, error) {
	s.gaugeData[name] = Gauge(value)
	return value, nil
}

//func (s *MemStorage) UpdateGauges(metrics []storage.GaugeMetric) ([]storage.GaugeMetric, error) {
//	for _, v := range metrics {
//		s.gaugeData[v.Name] = Gauge(v.Value)
//		//switch v.MType {
//		//case "gauge":
//		//	s.UpdateGauge(v.ID, *v.Value)
//		//case "counter":
//		//	s.UpdateCounter(v.ID, *v.Delta)
//
//	}
//	return metrics, nil
//}

func (s *MemStorage) GetCounterValue(id string) (int64, bool) {
	_, ok := s.counterData[id]
	//return val, ok
	return int64(s.counterData[id]), ok
}

func (s *MemStorage) GetGaugeValue(id string) (float64, bool) {
	_, ok := s.gaugeData[id]
	return float64(s.gaugeData[id]), ok
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

func (s *MemStorage) GetAllGauges() ([]storage.GaugeMetric, error) {
	gauges := make([]storage.GaugeMetric, 0, len(s.gaugeData))
	for name, value := range s.gaugeData {
		gauges = append(gauges, storage.GaugeMetric{Name: name, Value: float64(value)})
	}
	return gauges, nil
}

// GetAllCounters returns all counter metrics.
func (s *MemStorage) GetAllCounters() ([]storage.CounterMetric, error) {

	counters := make([]storage.CounterMetric, 0, len(s.counterData))
	for name, value := range s.counterData {
		counters = append(counters, storage.CounterMetric{Name: name, Value: int64(value)})
	}

	return counters, nil
}

func (s *MemStorage) BatchUpdate(w io.Writer, metrics []models.Metrics) error {
	for _, v := range metrics {
		switch v.MType {
		case "gauge":
			s.UpdateGauge(v.ID, *v.Value)
		case "counter":
			s.UpdateCounter(v.ID, *v.Delta)

		}
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(metrics[0]); err != nil {
		return fmt.Errorf("error occured on encoding result of batchupdate :%w", err)
	}

	return nil
}
