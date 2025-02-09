package metrics

import (
	"testing"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Sofja96/go-metrics.git/internal/agent/gzip"
	"github.com/Sofja96/go-metrics.git/internal/models"
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

func TestPrepareMetrics(t *testing.T) {
	collector := NewMetricsCollector()
	err := collector.GetMetrics()
	assert.NoError(t, err)

	err = func() error {
		collector.ValuesGauge["test_gauge"] = 123.45
		collector.ValuesCounter["test_counter"] = 10
		return nil
	}()
	assert.NoError(t, err)

	compressedData, err := collector.PrepareMetrics()
	assert.NoError(t, err, "Ошибка при подготовке метрик")
	assert.NotEmpty(t, compressedData, "Данные не должны быть пустыми")
}

func TestConvertToProtoMetrics(t *testing.T) {
	originalMetrics := []models.Metrics{
		{
			ID:    "metric1",
			MType: "counter",
			Delta: new(int64),
			Value: nil,
		},
		{
			ID:    "metric2",
			MType: "gauge",
			Delta: nil,
			Value: new(float64),
		},
	}

	compressedData, err := gzip.Compress(originalMetrics)
	assert.NoError(t, err, "Ошибка при сжатии данных")

	t.Run("Success", func(t *testing.T) {
		protoMetrics, err := ConvertToProtoMetrics(compressedData)

		assert.NoError(t, err)
		assert.Len(t, protoMetrics, len(originalMetrics))
		assert.Equal(t, protoMetrics[0].Id, originalMetrics[0].ID)
		assert.Equal(t, protoMetrics[1].Type, originalMetrics[1].MType)
	})

	t.Run("DecompressionError", func(t *testing.T) {
		invalidData := []byte("invalid json")
		protoMetrics, err := ConvertToProtoMetrics(invalidData)

		assert.Error(t, err)
		assert.Nil(t, protoMetrics)
		assert.Contains(t, err.Error(), "ошибка декомпрессии метрик")
	})
	t.Run("UnmarshalError", func(t *testing.T) {
		invalidJsonData := []byte("invalid json")
		compressedData, err := gzip.Compress(invalidJsonData)
		assert.NoError(t, err, "Ошибка при сжатии данных")

		protoMetrics, err := ConvertToProtoMetrics(compressedData)
		assert.Error(t, err)
		assert.Nil(t, protoMetrics)

		assert.Contains(t, err.Error(), "ошибка преобразования JSON в модель Metrics")
	})
}
