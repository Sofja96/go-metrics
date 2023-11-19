package metrics

import (
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"runtime"
)

var ValuesGauge = map[string]float64{}

var ValuesCounter = map[string]int64{}

func GetMetrics() []models.Metrics {

	var rtm runtime.MemStats
	// Read full mem stats
	runtime.ReadMemStats(&rtm)

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

	ValuesCounter["PollCount"]++

	return nil
}

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
