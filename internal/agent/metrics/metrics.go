package metrics

import (
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent/gzip"
	//стандартные библиотеки
	"math/rand"
	"runtime"

	//внешние библиотеки
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	//собственные модули
	"github.com/Sofja96/go-metrics.git/internal/models"
)

// Значения метрик типа gauge и counter.
type metrics struct {
	ValuesGauge   map[string]float64 // метрики типа gauge
	ValuesCounter map[string]int64   // метрики типа counter
}

// NewMetricsCollector - конструктор для создания экземпляра MetricsCollector.
func NewMetricsCollector() *metrics {
	return &metrics{
		ValuesGauge:   make(map[string]float64),
		ValuesCounter: make(map[string]int64),
	}
}

// GetMetrics - функция сбора метрик через runtime.MemStats а также случайного значения.
func (m *metrics) GetMetrics() error {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	// Заполняем значения метрик типа gauge.
	m.ValuesGauge["Alloc"] = float64(rtm.Alloc)
	m.ValuesGauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	m.ValuesGauge["Frees"] = float64(rtm.Frees)
	m.ValuesGauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	m.ValuesGauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	m.ValuesGauge["HeapIdle"] = float64(rtm.HeapIdle)
	m.ValuesGauge["HeapInuse"] = float64(rtm.HeapInuse)
	m.ValuesGauge["HeapObjects"] = float64(rtm.HeapObjects)
	m.ValuesGauge["HeapReleased"] = float64(rtm.HeapReleased)
	m.ValuesGauge["HeapSys"] = float64(rtm.HeapSys)
	m.ValuesGauge["LastGC"] = float64(rtm.LastGC)
	m.ValuesGauge["Lookups"] = float64(rtm.Lookups)
	m.ValuesGauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	m.ValuesGauge["MCacheSys"] = float64(rtm.MCacheSys)
	m.ValuesGauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	m.ValuesGauge["MSpanSys"] = float64(rtm.MSpanSys)
	m.ValuesGauge["Mallocs"] = float64(rtm.Mallocs)
	m.ValuesGauge["NextGC"] = float64(rtm.NextGC)
	m.ValuesGauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	m.ValuesGauge["NumGC"] = float64(rtm.NumGC)
	m.ValuesGauge["OtherSys"] = float64(rtm.OtherSys)
	m.ValuesGauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	m.ValuesGauge["StackInuse"] = float64(rtm.StackInuse)
	m.ValuesGauge["StackSys"] = float64(rtm.StackSys)
	m.ValuesGauge["Sys"] = float64(rtm.Sys)
	m.ValuesGauge["TotalAlloc"] = float64(rtm.TotalAlloc)
	m.ValuesGauge["GCSys"] = float64(rtm.GCSys)
	m.ValuesGauge["RandomValue"] = rand.Float64()

	// Увеличиваем счётчик PollCount.
	m.ValuesCounter["PollCount"]++

	return nil
}

// GetPSMetrics  - функция сбора метрик через gopsutil.
func (m *metrics) GetPSMetrics() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	m.ValuesGauge["TotalMemory"] = float64(v.Total)
	m.ValuesGauge["FreeMemory"] = float64(v.Free)
	cpu, _ := cpu.Percent(0, true)

	m.ValuesGauge["CPUutilization1"] = float64(cpu[0])

	return nil

}

// PrepareMetrics - преобразует собранные метрики в модели Metrics для отправки
func (m *metrics) PrepareMetrics() ([]byte, error) {
	allMetrics := make([]models.Metrics, 0, len(m.ValuesGauge)+len(m.ValuesCounter))

	for k, v := range m.ValuesGauge {
		val := v
		allMetrics = append(allMetrics, models.Metrics{
			MType: "gauge",
			ID:    k,
			Value: &val,
		})
	}

	for k, v := range m.ValuesCounter {
		val := v
		allMetrics = append(allMetrics, models.Metrics{
			MType: "counter",
			ID:    k,
			Delta: &val,
		})
	}

	compressedMetrics, err := gzip.Compress(allMetrics)
	if err != nil {
		return nil, fmt.Errorf("compression error %v", err)

	}

	return compressedMetrics, nil
}
