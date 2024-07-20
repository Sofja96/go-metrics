package memory

import (
	"bytes"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"testing"
)

func BenchmarkUpdateCounter(b *testing.B) {
	s, _ := NewMemStorage(0, "/tmp/metrics-db.json", false)
	for i := 0; i < b.N; i++ {
		s.UpdateCounter("test_counter", 1)
	}
}

func BenchmarkUpdateGauge(b *testing.B) {
	s, _ := NewMemStorage(0, "/tmp/metrics-db.json", false)
	for i := 0; i < b.N; i++ {
		s.UpdateGauge("test_gauge", 1.23)
	}
}

func BenchmarkGetCounterValue(b *testing.B) {
	s, _ := NewMemStorage(0, "/tmp/metrics-db.json", false)
	s.UpdateCounter("test_counter", 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetCounterValue("test_counter")
	}
}

func BenchmarkGetGaugeValue(b *testing.B) {
	s, _ := NewMemStorage(0, "/tmp/metrics-db.json", false)
	s.UpdateGauge("test_gauge", 1.23)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetGaugeValue("test_gauge")
	}
}

func BenchmarkBatchUpdate(b *testing.B) {
	s, _ := NewMemStorage(0, "/tmp/metrics-db.json", false)
	metrics := []models.Metrics{
		{MType: "gauge", ID: "test_gauge", Value: float64Ptr(1.23)},
		{MType: "counter", ID: "test_counter", Delta: int64Ptr(1)},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		s.BatchUpdate(&buf, metrics)
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
