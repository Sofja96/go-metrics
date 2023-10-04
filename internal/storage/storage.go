package storage

type Gauge float64
type Counter int64

type Storage interface {
	UpdateCounter(name string, value int64)
	UpdateGauge(name string, value float64)
	AllMetrics() *AllMetrics
	GetCounterValue(id string) (int64, bool)
	GetGaugeValue(id string) (float64, bool)
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
