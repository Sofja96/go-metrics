package storage

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
}

type CounterMetric struct {
	Name  string
	Value int64
}

type GaugeMetric struct {
	Name  string
	Value float64
}

type Metrics struct {
	Gauges   []GaugeMetric
	Counters []CounterMetric
}
