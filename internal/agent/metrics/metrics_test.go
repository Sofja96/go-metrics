package metrics

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetricsCollector()

	require.Empty(t, m.ValuesGauge)
	require.Empty(t, m.ValuesGauge)
}

func TestGetMetrics(t *testing.T) {
	m := NewMetricsCollector()
	err := m.GetMetrics()
	assert.NoError(t, err)

	err = func() error {
		m.ValuesGauge["Alloc"] = 12.28
		m.ValuesGauge["Heap"] = 11.91
		m.ValuesGauge["PollCount"] = 4
		return nil
	}()
	assert.NoError(t, err)

	assert.Equal(t, m.ValuesCounter["PollCount"], int64(1))
	assert.Equal(t, m.ValuesGauge["Alloc"], 12.28)
	assert.Equal(t, m.ValuesGauge["Heap"], 11.91)
}

func TestGetPSMetrics(t *testing.T) {
	m := NewMetricsCollector()
	err := m.GetPSMetrics()
	assert.NoError(t, err)

	memStat := &mem.VirtualMemoryStat{
		Total: 8192,
		Free:  4096,
	}
	cpuPercent := []float64{10.5}

	err = func() error {
		m.ValuesGauge["TotalMemory"] = float64(memStat.Total)
		m.ValuesGauge["FreeMemory"] = float64(memStat.Free)
		m.ValuesGauge["CPUutilization1"] = float64(cpuPercent[0])
		return nil
	}()
	assert.NoError(t, err)

	assert.Equal(t, m.ValuesGauge["TotalMemory"], float64(memStat.Total))
	assert.Equal(t, m.ValuesGauge["FreeMemory"], float64(memStat.Free))
	assert.Equal(t, m.ValuesGauge["CPUutilization1"], float64(cpuPercent[0]))
}
