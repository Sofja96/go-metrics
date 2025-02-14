package database

import (
	"context"
	"testing"

	"github.com/Sofja96/go-metrics.git/internal/models"
)

// Функция для создания нового подключения к базе данных для тестов
func setupDB(b *testing.B) *Postgres {
	dsn := "postgres://metrics:userpassword@localhost:5432/metrics"
	db, err := NewStorage(context.Background(), dsn)
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func BenchmarkUpdateCounter(b *testing.B) {
	db := setupDB(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.UpdateCounter(context.Background(), "test_counter", 1)
		if err != nil {
			b.Fatalf("Failed to update counter: %v", err)
		}
	}
}

func BenchmarkUpdateGauge(b *testing.B) {
	db := setupDB(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.UpdateGauge(context.Background(), "test_gauge", 1.23)
		if err != nil {
			b.Fatalf("Failed to update gauge: %v", err)
		}
	}
}

func BenchmarkGetCounterValue(b *testing.B) {
	db := setupDB(b)
	_, err := db.UpdateCounter(context.Background(), "test_counter", 1)
	if err != nil {
		b.Fatalf("Failed to setup counter: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := db.GetCounterValue(context.Background(), "test_counter")
		if !ok {
			b.Fatalf("Failed to get counter value")
		}
	}
}

func BenchmarkGetGaugeValue(b *testing.B) {
	db := setupDB(b)
	_, err := db.UpdateGauge(context.Background(), "test_gauge", 1.23)
	if err != nil {
		b.Fatalf("Failed to setup gauge: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := db.GetGaugeValue(context.Background(), "test_gauge")
		if !ok {
			b.Fatalf("Failed to get gauge value")
		}
	}
}

func BenchmarkBatchUpdate(b *testing.B) {
	db := setupDB(b)
	metrics := []models.Metrics{
		{MType: "gauge", ID: "test_gauge", Value: float64Ptr(1.23)},
		{MType: "counter", ID: "test_counter", Delta: int64Ptr(1)},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.BatchUpdate(context.Background(), metrics)
		if err != nil {
			b.Fatalf("Failed to batch update: %v", err)
		}
	}
}

func float64Ptr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
