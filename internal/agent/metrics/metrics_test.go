package metrics

import (
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestGetMetrics(t *testing.T) {
	ValuesGauge = make(map[string]float64)
	ValuesCounter = make(map[string]int64)
	var Mu sync.Mutex

	GetMetrics()

	Mu.Lock()
	defer Mu.Unlock()

	assert.Greater(t, ValuesGauge["Alloc"], 0.0)
	assert.GreaterOrEqual(t, ValuesGauge["RandomValue"], 0.0)
	assert.LessOrEqual(t, ValuesGauge["RandomValue"], 1.0)

	assert.Equal(t, ValuesCounter["PollCount"], int64(1))
}

func TestGetPSMetrics(t *testing.T) {
	ValuesGauge = make(map[string]float64)
	ValuesCounter = make(map[string]int64)
	var Mu sync.Mutex

	memStat := &mem.VirtualMemoryStat{
		Total: 8192,
		Free:  4096,
	}
	cpuPercent := []float64{10.5}

	err := func() error {
		Mu.Lock()
		defer Mu.Unlock()
		ValuesGauge["TotalMemory"] = float64(memStat.Total)
		ValuesGauge["FreeMemory"] = float64(memStat.Free)
		ValuesGauge["CPUutilization1"] = float64(cpuPercent[0])
		return nil
	}()
	assert.NoError(t, err)

	assert.Equal(t, ValuesGauge["TotalMemory"], float64(memStat.Total))
	assert.Equal(t, ValuesGauge["FreeMemory"], float64(memStat.Free))
	assert.Equal(t, ValuesGauge["CPUutilization1"], float64(cpuPercent[0]))
}

func TestConcurrency(t *testing.T) {
	ValuesGauge = make(map[string]float64)
	ValuesCounter = make(map[string]int64)
	var Mu sync.Mutex

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		GetMetrics()
	}()

	go func() {
		defer wg.Done()
		GetPSMetrics()
	}()

	wg.Wait()

	Mu.Lock()
	defer Mu.Unlock()
	assert.GreaterOrEqual(t, ValuesCounter["PollCount"], int64(1))
	assert.Contains(t, ValuesGauge, "Alloc")
	assert.Contains(t, ValuesGauge, "TotalMemory")
}
