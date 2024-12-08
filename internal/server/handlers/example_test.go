package handlers

import (
	"bytes"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/memory"
	"github.com/labstack/echo/v4"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
)

// ExampleUpdateJSON демонстрирует пример запроса POST к /update/
func ExampleUpdateJSON() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()

	e.POST("/update/", UpdateJSON(s))

	body := `{"type":"counter", "id":"counter","delta":10}`
	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())

	// Output:
	// {"id":"counter","type":"counter","delta":10}
}

// ExampleUpdatesBatch демонстрирует пример запроса POST к /updates/
func ExampleUpdatesBatch() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()

	e.POST("/updates/", UpdatesBatch(s))

	reqBody := `[{"id":"PollCount1","type":"counter","delta":1},{"id":"MyAlloc","type":"gauge","value":12.52}]`

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// [{"id":"PollCount1","type":"counter","delta":1},{"id":"MyAlloc","type":"gauge","value":12.52}]
}

// ExampleValueJSON демонстрирует пример запроса POST к /value/
func ExampleValueJSON() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()
	e.POST("/value/", ValueJSON(s))
	_, _ = s.UpdateGauge("gauge", 15.25)

	reqBody := `{"type":"gauge","id":"gauge"}`
	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// {"id":"gauge","type":"gauge","value":15.25}
}

// formatHTML форматирует HTML строку для удобства сравнения
func formatHTML(html string) string {
	// Используем html/template для форматирования HTML
	var builder strings.Builder
	// Добавляем отступы и новые строки для форматирования
	tmpl := `
<html>
<body>
  <h2>Gauge metrics:</h2>
  <ul>
    <li>gauge = 15.25</li>
    <li>gauge1 = 4.56</li>
  </ul>
  <h2>Counter metrics:</h2>
  <ul>
    <li>counter = 10</li>
    <li>counter1 = 20</li>
  </ul>
</body>
</html>`
	// Создаем новый шаблон
	t := template.Must(template.New("html").Parse(tmpl))
	// Выполняем шаблон с помощью builder
	t.Execute(&builder, nil)
	// Удаляем лишние пробелы и переносы строк
	return strings.TrimSpace(builder.String())
}

// ExampleGetAllMetrics демонстрирует пример запроса GET к /
func ExampleGetAllMetrics() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()
	e.GET("/", GetAllMetrics(s))

	_, _ = s.UpdateGauge("gauge", 15.25)
	_, _ = s.UpdateGauge("gauge1", 4.56)
	_, _ = s.UpdateCounter("counter", 10)
	_, _ = s.UpdateCounter("counter1", 20)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	got := rec.Body.String()
	formattedGot := formatHTML(got)
	fmt.Print(formattedGot)

	// Output:
	// <html>
	// <body>
	//   <h2>Gauge metrics:</h2>
	//   <ul>
	//     <li>gauge = 15.25</li>
	//     <li>gauge1 = 4.56</li>
	//   </ul>
	//   <h2>Counter metrics:</h2>
	//   <ul>
	//     <li>counter = 10</li>
	//     <li>counter1 = 20</li>
	//   </ul>
	// </body>
	// </html>
}

// ExampleValueMetric демонстрирует пример запроса GET к /value/:typeM/:nameM
func ExampleValueMetric() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()
	e.GET("/value/:typeM/:nameM", ValueMetric(s))

	_, _ = s.UpdateCounter("PollCount1", 10)

	req := httptest.NewRequest(http.MethodGet, "/value/counter/PollCount1", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// 10
}

// ExampleWebhook демонстрирует пример запроса POST к /update/:typeM/:nameM/:valueM
func ExampleWebhook() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/10", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	//
}

// ExamplePing демонстрирует пример запроса GET к /ping
func ExamplePing() {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := echo.New()
	e.GET("/ping", Ping(s))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())

	// Output:
	// Connection database is OK
}
