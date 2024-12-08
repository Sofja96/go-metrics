package memory

import (
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	storagemock "github.com/Sofja96/go-metrics.git/internal/server/storage/mocks"
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCounter(t *testing.T) {
	s, _ := NewMemStorage(300, "/tmp/metrics-db.json", false)
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
			counter, err := s.UpdateCounter(test.metricsName, test.value)
			assert.NoError(t, err)
			assert.Equal(t, test.result, counter)
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	s, _ := NewMemStorage(300, "/tmp/metrics-db.json", false)
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
			gauge, err := s.UpdateGauge(test.metricsName, test.value)
			assert.NoError(t, err)
			assert.Equal(t, test.result, gauge)
		})
	}
}

func TestPing(t *testing.T) {
	s, _ := NewMemStorage(0, "", false)
	err := s.Ping()
	assert.NoError(t, err, "Ping should not return an error")
}

func TestGetCounterValue(t *testing.T) {
	s, _ := NewMemStorage(300, "", false)
	_, err := s.UpdateCounter("testCounter", 10)
	assert.NoError(t, err)

	value, ok := s.GetCounterValue("testCounter")
	assert.True(t, ok, "GetCounterValue should return true for existing counter")
	assert.Equal(t, int64(10), value)

	_, ok = s.GetCounterValue("nonExistentCounter")
	assert.False(t, ok, "GetCounterValue should return false for non-existing counter")
}

func TestGetGaugeValue(t *testing.T) {
	s, _ := NewMemStorage(300, "", false)
	_, err := s.UpdateGauge("testGauge", 20.5)
	assert.NoError(t, err)

	value, ok := s.GetGaugeValue("testGauge")
	assert.True(t, ok, "GetGaugeValue should return true for existing gauge")
	assert.Equal(t, 20.5, value)

	_, ok = s.GetGaugeValue("nonExistentGauge")
	assert.False(t, ok, "GetGaugeValue should return false for non-existing gauge")
}

func TestGetAllCounters(t *testing.T) {
	s, _ := NewMemStorage(300, "", false)
	_, err := s.UpdateCounter("testCounter1", 5)
	assert.NoError(t, err)
	_, err = s.UpdateCounter("testCounter2", 15)
	assert.NoError(t, err)

	counters, err := s.GetAllCounters()
	assert.NoError(t, err)
	assert.Len(t, counters, 2)
}

func TestGetAllGauges(t *testing.T) {
	s, _ := NewMemStorage(300, "", false)
	_, err := s.UpdateGauge("testGauge1", 10.5)
	assert.NoError(t, err)
	_, err = s.UpdateGauge("testGauge2", 20.5)
	assert.NoError(t, err)

	gauges, err := s.GetAllGauges()
	assert.NoError(t, err)
	assert.Len(t, gauges, 2)
}

func TestGetAllMetrics(t *testing.T) {
	s, _ := NewMemStorage(300, "", false)
	_, err := s.UpdateCounter("counter1", 1)
	assert.NoError(t, err)
	_, err = s.UpdateGauge("gauge1", 2.2)
	assert.NoError(t, err)

	allMetrics := s.AllMetrics()
	assert.Len(t, allMetrics.Counter, 1)
	assert.Len(t, allMetrics.Gauge, 1)
	assert.Equal(t, Gauge(2.2), allMetrics.Gauge["gauge1"])
	assert.Equal(t, Counter(1), allMetrics.Counter["counter1"])
}

type mocks struct {
	storage *storagemock.MockStorage
}

func TestBatchUpdate(t *testing.T) {
	s, _ := NewMemStorage(300, "", false)

	metrics := []models.Metrics{
		{MType: "gauge", ID: "gauge1", Value: floatPtr(10.5)},
		{MType: "counter", ID: "counter1", Delta: intPtr(5)},
	}

	err := s.BatchUpdate(metrics)
	assert.NoError(t, err)

	gaugeValue, _ := s.GetGaugeValue("gauge1")
	assert.Equal(t, 10.5, gaugeValue)

	counterValue, _ := s.GetCounterValue("counter1")
	assert.Equal(t, int64(5), counterValue)
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
				m.storage.EXPECT().BatchUpdate(args.metrics)
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
				m.storage.EXPECT().BatchUpdate(args.metrics).Return(fmt.Errorf("error batch update"))
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
			err := m.storage.BatchUpdate(tt.args.metrics)

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
