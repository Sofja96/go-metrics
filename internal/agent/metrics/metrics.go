package metrics

import (
	//стандартные библиотеки
	"math/rand"
	"runtime"
	"sync"

	//внешние библиотеки
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	//собственные модули
	"github.com/Sofja96/go-metrics.git/internal/models"
)

// Значения метрик типа gauge и counter.
var (
	ValuesGauge   = map[string]float64{} // метрики типа gauge
	ValuesCounter = map[string]int64{}   // метрики типа counter
	Mu            sync.Mutex
)

// GetMetrics - функция сбора метрик через runtime.MemStats а также случайного значения.
func GetMetrics() []models.Metrics {

	Mu.Lock()
	defer Mu.Unlock()

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	// Заполняем значения метрик типа gauge.
	ValuesGauge["Alloc"] = float64(rtm.Alloc)
	ValuesGauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	ValuesGauge["Frees"] = float64(rtm.Frees)
	ValuesGauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	ValuesGauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	ValuesGauge["HeapIdle"] = float64(rtm.HeapIdle)
	ValuesGauge["HeapInuse"] = float64(rtm.HeapInuse)
	ValuesGauge["HeapObjects"] = float64(rtm.HeapObjects)
	ValuesGauge["HeapReleased"] = float64(rtm.HeapReleased)
	ValuesGauge["HeapSys"] = float64(rtm.HeapSys)
	ValuesGauge["LastGC"] = float64(rtm.LastGC)
	ValuesGauge["Lookups"] = float64(rtm.Lookups)
	ValuesGauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	ValuesGauge["MCacheSys"] = float64(rtm.MCacheSys)
	ValuesGauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	ValuesGauge["MSpanSys"] = float64(rtm.MSpanSys)
	ValuesGauge["Mallocs"] = float64(rtm.Mallocs)
	ValuesGauge["NextGC"] = float64(rtm.NextGC)
	ValuesGauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	ValuesGauge["NumGC"] = float64(rtm.NumGC)
	ValuesGauge["OtherSys"] = float64(rtm.OtherSys)
	ValuesGauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	ValuesGauge["StackInuse"] = float64(rtm.StackInuse)
	ValuesGauge["StackSys"] = float64(rtm.StackSys)
	ValuesGauge["Sys"] = float64(rtm.Sys)
	ValuesGauge["TotalAlloc"] = float64(rtm.TotalAlloc)
	ValuesGauge["GCSys"] = float64(rtm.GCSys)
	ValuesGauge["RandomValue"] = rand.Float64()

	// Увеличиваем счётчик PollCount.
	ValuesCounter["PollCount"]++

	return nil
}

// GetPSMetrics  - функция сбора метрик через gopsutil.
func GetPSMetrics() ([]models.Metrics, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	ValuesGauge["TotalMemory"] = float64(v.Total)
	ValuesGauge["FreeMemory"] = float64(v.Free)
	cpu, _ := cpu.Percent(0, true)

	ValuesGauge["CPUutilization1"] = float64(cpu[0])

	return nil, nil

}
