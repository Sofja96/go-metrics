package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/Sofja96/go-metrics.git/internal/server/storage/memory"
)

func BenchmarkWebhook(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := Webhook(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/counter/metric1/100", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}

func BenchmarkUpdateJSON(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := UpdateJSON(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBufferString(`{"id":"metric1","m_type":"counter","delta":100}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}

func BenchmarkUpdatesBatch(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := UpdatesBatch(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/batchupdate", bytes.NewBufferString(`[{"id":"metric1","m_type":"counter","delta":100}]`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}

func BenchmarkValueMetric(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := ValueMetric(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/value/counter/metric1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/value/:typeM/:nameM")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}

func BenchmarkValueJSON(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := ValueJSON(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/valuejson", bytes.NewBufferString(`{"id":"metric1","m_type":"counter"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}

func BenchmarkGetAllMetrics(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := GetAllMetrics(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/allmetrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}

func BenchmarkPing(b *testing.B) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	handler := Ping(s)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler(c)
	}
}
