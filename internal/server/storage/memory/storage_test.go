package memory

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/Sofja96/go-metrics.git/internal/models"
	storagemock "github.com/Sofja96/go-metrics.git/internal/server/storage/mocks"
)

func TestUpdateCounter(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "/tmp/metrics-db.json", false)
	testCases := []struct {
		name        string
		metricsName string
		value       int64
		result      int64
	}{
		{name: "UpdateCounter() Test 1", metricsName: "testCounter1", value: 1, result: 1},
		{name: "UpdateCounter() Test 2", metricsName: "testCounter2", value: 1, result: 1},
		{name: "UpdateCounter() Test 3", metricsName: "testCounter1", value: 1, result: 2},
		{name: "UpdateCounter() Test 4", metricsName: "testCounter1", value: 10000, result: 10002},
		{name: "UpdateCounter() Test 5", metricsName: "testCounter2", value: 0, result: 1},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			counter, err := s.UpdateCounter(ctx, test.metricsName, test.value)
			assert.NoError(t, err)
			assert.Equal(t, test.result, counter)
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "/tmp/metrics-db.json", false)
	testCases := []struct {
		name        string
		metricsName string
		value       float64
		result      float64
	}{
		{name: "UpdateGauge() Test 1", metricsName: "testGauge1", value: 1, result: 1.0},
		{name: "UpdateGauge() Test 2", metricsName: "testGauge2", value: 1.0, result: 1.0},
		{name: "UpdateGauge() Test 4", metricsName: "testGauge1", value: 10000, result: 10000.0},
		{name: "UpdateGauge() Test 5", metricsName: "testGauge2", value: 0, result: 0.0},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			gauge, err := s.UpdateGauge(ctx, test.metricsName, test.value)
			assert.NoError(t, err)
			assert.Equal(t, test.result, gauge)
		})
	}
}

func TestPing(t *testing.T) {
	s, _ := NewMemStorage(context.Background(), 0, "", false)
	err := s.Ping(context.Background())
	assert.NoError(t, err, "Ping should not return an error")
}

func TestGetCounterValue(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)
	_, err := s.UpdateCounter(ctx, "testCounter", 10)
	assert.NoError(t, err)

	value, ok := s.GetCounterValue(ctx, "testCounter")
	assert.True(t, ok, "GetCounterValue should return true for existing counter")
	assert.Equal(t, int64(10), value)

	_, ok = s.GetCounterValue(ctx, "nonExistentCounter")
	assert.False(t, ok, "GetCounterValue should return false for non-existing counter")
}

func TestGetGaugeValue(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)
	_, err := s.UpdateGauge(ctx, "testGauge", 20.5)
	assert.NoError(t, err)

	value, ok := s.GetGaugeValue(ctx, "testGauge")
	assert.True(t, ok, "GetGaugeValue should return true for existing gauge")
	assert.Equal(t, 20.5, value)

	_, ok = s.GetGaugeValue(ctx, "nonExistentGauge")
	assert.False(t, ok, "GetGaugeValue should return false for non-existing gauge")
}

func TestGetAllCounters(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)
	_, err := s.UpdateCounter(ctx, "testCounter1", 5)
	assert.NoError(t, err)
	_, err = s.UpdateCounter(ctx, "testCounter2", 15)
	assert.NoError(t, err)

	counters, err := s.GetAllCounters(ctx)
	assert.NoError(t, err)
	assert.Len(t, counters, 2)
}

func TestGetAllGauges(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)
	_, err := s.UpdateGauge(ctx, "testGauge1", 10.5)
	assert.NoError(t, err)
	_, err = s.UpdateGauge(ctx, "testGauge2", 20.5)
	assert.NoError(t, err)

	gauges, err := s.GetAllGauges(ctx)
	assert.NoError(t, err)
	assert.Len(t, gauges, 2)
}

func TestGetAllMetrics(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)
	_, err := s.UpdateCounter(ctx, "counter1", 1)
	assert.NoError(t, err)
	_, err = s.UpdateGauge(ctx, "gauge1", 2.2)
	assert.NoError(t, err)

	allMetrics := s.AllMetrics(ctx)
	assert.Len(t, allMetrics.Counter, 1)
	assert.Len(t, allMetrics.Gauge, 1)
	assert.Equal(t, Gauge(2.2), allMetrics.Gauge["gauge1"])
	assert.Equal(t, Counter(1), allMetrics.Counter["counter1"])
}

type mocks struct {
	storage *storagemock.MockStorage
}

func TestUpdateGaugeData(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)

	testCases := []struct {
		name      string
		inputData map[string]Gauge
		expected  map[string]Gauge
	}{
		{
			name:      "UpdateGaugeData() Test 1",
			inputData: map[string]Gauge{"gauge1": 10.5, "gauge2": 20.0},
			expected:  map[string]Gauge{"gauge1": 10.5, "gauge2": 20.0},
		},
		{
			name:      "UpdateGaugeData() Test 2",
			inputData: map[string]Gauge{"gauge3": 15.5},
			expected:  map[string]Gauge{"gauge3": 15.5},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateGaugeData(ctx, test.inputData)

			s.mutex.Lock()
			defer s.mutex.Unlock()

			assert.Equal(t, test.expected, s.gaugeData)
		})
	}
}

func TestUpdateCounterData(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)

	testCases := []struct {
		name      string
		inputData map[string]Counter
		expected  map[string]Counter
	}{
		{
			name:      "UpdateCounterData() Test 1",
			inputData: map[string]Counter{"counter1": 5, "counter2": 10},
			expected:  map[string]Counter{"counter1": 5, "counter2": 10},
		},
		{
			name:      "UpdateCounterData() Test 2",
			inputData: map[string]Counter{"counter3": 15},
			expected:  map[string]Counter{"counter3": 15},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateCounterData(ctx, test.inputData)

			s.mutex.Lock()
			defer s.mutex.Unlock()

			assert.Equal(t, test.expected, s.counterData)
		})
	}
}

func TestBatchUpdate(t *testing.T) {
	ctx := context.Background()
	s, _ := NewMemStorage(ctx, 300, "", false)

	t.Run("BatchUpdate_SUCCESS", func(t *testing.T) {
		metrics := []models.Metrics{
			{MType: "gauge", ID: "gauge1", Value: floatPtr(10.5)},
			{MType: "counter", ID: "counter1", Delta: intPtr(5)},
		}

		err := s.BatchUpdate(ctx, metrics)
		assert.NoError(t, err)

		gaugeValue, _ := s.GetGaugeValue(ctx, "gauge1")
		assert.Equal(t, 10.5, gaugeValue)

		counterValue, _ := s.GetCounterValue(ctx, "counter1")
		assert.Equal(t, int64(5), counterValue)
	})

	t.Run("BatchUpdate_ERROR", func(t *testing.T) {
		metrics := []models.Metrics{
			{MType: "gauge1", ID: "gauge1", Value: floatPtr(10.5)},
			{MType: "counter1", ID: "counter1", Delta: intPtr(5)},
		}
		err := s.BatchUpdate(ctx, metrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported metrics type")

	})
}

func TestBatchUpdate_Mock(t *testing.T) {
	type (
		args struct {
			metrics []models.Metrics
		}
		mockBehavior func(m *mocks, args args)
	)
	tests := []struct {
		name         string
		args         args
		mockBehavior mockBehavior
		wantErr      bool
	}{
		{
			name: "Update batch Success",
			args: args{metrics: []models.Metrics{
				{MType: "gauge", ID: "gauge1", Value: floatPtr(10.5)},
				{MType: "counter", ID: "counter1", Delta: intPtr(5)},
			}},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().BatchUpdate(gomock.Any(), args.metrics)
			},
			wantErr: false,
		},
		{
			name: "Update batch Error",
			args: args{metrics: []models.Metrics{
				{MType: "gauge", ID: "gauge1", Value: floatPtr(10.5)},
				{MType: "counter", ID: "counter1", Delta: intPtr(5)},
			}},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().BatchUpdate(gomock.Any(), args.metrics).Return(fmt.Errorf("error batch update"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
			}

			tt.mockBehavior(m, tt.args)
			err := m.storage.BatchUpdate(context.Background(), tt.args.metrics)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})

	}

}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}

func TestNewMemStorage(t *testing.T) {
	ctx := context.Background()

	t.Run("NewMemStorage with restore", func(t *testing.T) {
		file, err := os.CreateTemp("", "metrics-db-test.json")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		testData := `{"counter":{"testCounter":10},"gauge":{"testGauge":20.5}}`
		_, err = file.WriteString(testData)
		assert.NoError(t, err)
		file.Close()

		s, err := NewMemStorage(ctx, 0, file.Name(), true)
		assert.NoError(t, err)

		counterValue, ok := s.GetCounterValue(ctx, "testCounter")
		assert.True(t, ok)
		assert.Equal(t, int64(10), counterValue)

		gaugeValue, ok := s.GetGaugeValue(ctx, "testGauge")
		assert.True(t, ok)
		assert.Equal(t, 20.5, gaugeValue)
	})

	t.Run("NewMemStorage with restore check error", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "metrics-db-test.json")
		assert.NoError(t, err)

		_, err = tempFile.WriteString("{invalid_json}")
		assert.NoError(t, err)
		tempFile.Close()

		s, err := NewMemStorage(ctx, 1, tempFile.Name(), true)
		assert.Error(t, err, "failed to restore data from file")
		assert.Nil(t, s)

		os.Remove(tempFile.Name())
		s, err = NewMemStorage(ctx, 1, tempFile.Name(), true)
		assert.NoError(t, err, "Storage should handle missing file without error")
		assert.NotNil(t, s)
	})

	t.Run("NewMemStorage without restore", func(t *testing.T) {
		s, err := NewMemStorage(ctx, 0, "", false)
		assert.NoError(t, err)

		counters, err := s.GetAllCounters(ctx)
		assert.NoError(t, err)
		assert.Empty(t, counters)

		gauges, err := s.GetAllGauges(ctx)
		assert.NoError(t, err)
		assert.Empty(t, gauges)
	})
}

func TestNewInMemStorage(t *testing.T) {
	ctx := context.Background()

	t.Run("NewInMemStorage with valid parameters", func(t *testing.T) {
		file, err := os.CreateTemp("", "metrics-db-test.json")
		assert.NoError(t, err)
		defer os.Remove(file.Name())

		s, err := NewInMemStorage(ctx, 0, file.Name(), false)
		assert.NoError(t, err)

		assert.NotNil(t, s)

		_, ok := s.(*MemStorage)
		assert.True(t, ok)
	})

	t.Run("NewInMemStorage with invalid file path", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "metrics-db-test.json")
		assert.NoError(t, err)

		_, err = tempFile.WriteString("{invalid_json}")
		assert.NoError(t, err)
		tempFile.Close()

		s, err := NewInMemStorage(ctx, 0, tempFile.Name(), true)

		assert.Error(t, err)
		assert.Nil(t, s)
	})
}
