package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/memory"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebhook(t *testing.T) {
	s, err := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	require.NoError(t, err)
	e := CreateServer(s)
	httpTestServer := httptest.NewServer(e)
	defer httpTestServer.Close()

	type result struct {
		code int
		body string
	}

	tt := []struct {
		name     string
		path     string
		expected result
	}{
		{
			name: "Push counter",
			path: fmt.Sprintf("%s/update/counter/PollCount/10", httpTestServer.URL),
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Push gauge",
			path: fmt.Sprintf("%s/update/gauge/Alloc/13.123", httpTestServer.URL),
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Push unknown metric kind",
			path: fmt.Sprintf("%s/update/unknown/Alloc/12.123", httpTestServer.URL),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push without name metric",
			path: fmt.Sprintf("%s/update/Alloc/12.123", httpTestServer.URL),
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Push counter with invalid name",
			path: fmt.Sprintf("%s/update/counter/Alloc/18446744073709551617", httpTestServer.URL),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push counter with invalid value",
			path: fmt.Sprintf("%s/update/gauge/PollCount/10\\.0", httpTestServer.URL),

			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push method get",
			path: fmt.Sprintf("%s/", httpTestServer.URL),

			expected: result{
				code: http.StatusMethodNotAllowed,
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			tr := &http.Transport{}
			client := &http.Client{Transport: tr}

			res, err := client.Post(tc.path, "text/plain", nil)
			require.NoError(t, err)

			assert.Equal(tc.expected.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}

func TestValueMetric(t *testing.T) {
	s, err := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	require.NoError(t, err)
	_, err = s.UpdateCounter("PollCount1", 10)
	require.NoError(t, err)
	_, err = s.UpdateGauge("Alloc1", 10.25)
	require.NoError(t, err)
	e := CreateServer(s)
	httpTestServer := httptest.NewServer(e)
	defer httpTestServer.Close()
	type result struct {
		code int
		body string
	}

	tt := []struct {
		name     string
		path     string
		expected result
	}{
		{
			name: "get counter",
			path: fmt.Sprintf("%s/value/counter/PollCount1", httpTestServer.URL),
			expected: result{
				code: http.StatusOK,
				body: "10",
			},
		},
		{
			name: "Get gauge",
			path: fmt.Sprintf("%s/value/gauge/Alloc1", httpTestServer.URL),
			expected: result{
				code: http.StatusOK,
				body: "10.25",
			},
		},
		{
			name: "Get unknown metric kind",
			path: fmt.Sprintf("%s/value/unknown/Alloc", httpTestServer.URL),
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get unknown counter",
			path: fmt.Sprintf("%s/value/counter/unknown", httpTestServer.URL),
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get unknown gauge",
			path: fmt.Sprintf("%s/value/gauge/unknown", httpTestServer.URL),
			expected: result{
				code: http.StatusNotFound,
			},
		},
	}

	for _, tc := range tt {
		assert := assert.New(t)
		t.Run(tc.name, func(t *testing.T) {
			tr := &http.Transport{}
			client := &http.Client{Transport: tr}

			res, err := client.Get(tc.path)
			require.NoError(t, err)

			assert.Equal(tc.expected.code, res.StatusCode)

			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.NotEmpty(string(respBody))
				assert.Equal(tc.expected.body, string(respBody))

				defer res.Body.Close()
			}
		})
	}
}

func TestUpdatesBatch(t *testing.T) {
	s, _ := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	e := CreateServer(s)
	httpTestServer := httptest.NewServer(e)
	defer httpTestServer.Close()

	type result struct {
		code int
		body string
	}

	delta := int64(rand.Intn(10))
	value := rand.Float64() * 10

	metrics := []models.Metrics{
		{MType: "counter", ID: "BatchCounter1", Delta: &delta},
		{MType: "gauge", ID: "BatchGauge1", Value: &value},
	}

	body, _ := json.Marshal(metrics)

	tt := []struct {
		name     string
		path     string
		body     []byte
		expected result
	}{
		{
			name: "Batch update",
			path: fmt.Sprintf("%s/updates/", httpTestServer.URL),
			body: body,
			expected: result{
				code: http.StatusOK,
				body: string(body),
			},
		},
		{
			name: "Empty metrics array",
			path: fmt.Sprintf("%s/updates/", httpTestServer.URL),
			body: []byte("[]"),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Invalid JSON format",
			path: fmt.Sprintf("%s/updates/", httpTestServer.URL),
			body: []byte("invalid_json"),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		assert := assert.New(t)
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodPost, tc.path, bytes.NewBuffer(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			e.ServeHTTP(recorder, req)

			assert.Equal(tc.expected.code, recorder.Code)

			if tc.expected.code == http.StatusOK {
				respBody := recorder.Body.Bytes()
				contentEncoding := recorder.Header().Get("Content-Encoding")

				if contentEncoding == "gzip" {
					gzipReader, err := gzip.NewReader(bytes.NewReader(respBody))
					require.NoError(t, err)
					defer gzipReader.Close()
					respBody, err = io.ReadAll(gzipReader)
					require.NoError(t, err)
				}

				var responseData []models.Metrics
				err = json.Unmarshal(respBody, &responseData)
				require.NoError(t, err)

				assert.Equal(strings.TrimSpace(tc.expected.body), strings.TrimSpace(string(respBody)))

				log.Println("Request body: ", string(tc.body))
				log.Println("Response body: ", string(respBody))

			}
		})
	}
}

func TestValueJSON(t *testing.T) {
	s, err := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	require.NoError(t, err)
	_, err = s.UpdateCounter("counter", 10)
	require.NoError(t, err)
	_, err = s.UpdateGauge("gauge", 15.25)
	require.NoError(t, err)

	e := CreateServer(s)
	httpTestServer := httptest.NewServer(e)
	defer httpTestServer.Close()

	validCounterJSON := `{"type": "counter", "id": "counter"}`
	validCounterJSONResp := `{"id":"counter","type":"counter","delta":10}`

	validGagugeJSON := `{"type": "gauge", "id": "gauge"}`
	validGaugeJSONResp := `{"id":"gauge","type":"gauge","value":15.25}`

	notValidCounterJSON := `{"type": "", "id": "counter"}`

	notValidGaugeJSON := `{"type": "gauge", "id": ""}`

	invalidJson := `{"type": "gauge", "id": gauge}`

	noCounterJSON := `{"type": "counter", "id": "counter1"}`
	noGagugeJSON := `{"type": "gauge", "id": "gauge1"}`

	type result struct {
		code int
		body string
	}

	tt := []struct {
		name        string
		path        string
		body        string
		expected    result
		contentType string
	}{
		{
			name:        "get counter1",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        validCounterJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusOK,
				body: validCounterJSONResp,
			},
		},
		{
			name:        "Get gauge",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        validGagugeJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusOK,
				body: validGaugeJSONResp,
			},
		},
		{
			name:        "Get unknown metric kind",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        notValidCounterJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "Get empty id metric",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        notValidGaugeJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "Get metrics with wrong content type",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        validGagugeJSON,
			contentType: "text/plain",
			expected: result{
				code: http.StatusUnsupportedMediaType,
			},
		},
		{
			name:        "Get metrics with invalid Json",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        invalidJson,
			contentType: "application/json",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "Get not exists metrics Counter",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        noCounterJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "Get not exists metrics Gauge",
			path:        fmt.Sprintf("%s/value/", httpTestServer.URL),
			body:        noGagugeJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusNotFound,
			},
		},
	}

	for _, tc := range tt {
		assert := assert.New(t)
		t.Run(tc.name, func(t *testing.T) {
			tr := &http.Transport{}
			client := &http.Client{Transport: tr}
			res, err := client.Post(tc.path, tc.contentType, strings.NewReader(tc.body))
			require.NoError(t, err)
			assert.Equal(tc.expected.code, res.StatusCode)
			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.NotEmpty(string(respBody))
				assert.Equal(strings.TrimSpace(tc.expected.body), strings.TrimSpace(string(respBody)))

				defer res.Body.Close()

				log.Println("Request body: ", tc.body)
				log.Println("Expected body: ", tc.expected.body)
				log.Println("Response body: ", string(respBody))
			}
		})
	}
}

func TestUpdateJSON(t *testing.T) {
	s, err := memory.NewMemStorage(300, "/tmp/metrics-db.json", false)
	require.NoError(t, err)

	e := CreateServer(s)
	httpTestServer := httptest.NewServer(e)
	defer httpTestServer.Close()

	validCounterJSON := `{"type":"counter", "id":"counter","delta":10}`
	validCounterJSONResp := `{"id":"counter","type":"counter","delta":10}`

	validGagugeJSON := `{"type": "gauge", "id": "gauge", "value":15.25}`
	validGaugeJSONResp := `{"id":"gauge","type":"gauge","value":15.25}`

	notValidCounterJSON := `{"type": "", "id": "counter"}` //404

	notValidGaugeJSON := `{"type": "gauge", "id": ""}` //404

	invalidJson := `{"type": "gauge", "id": gauge}` //400

	type result struct {
		code int
		body string
	}

	tt := []struct {
		name        string
		path        string
		body        string
		expected    result
		contentType string
	}{
		{
			name:        "update counter",
			path:        fmt.Sprintf("%s/update/", httpTestServer.URL),
			body:        validCounterJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusOK,
				body: validCounterJSONResp,
			},
		},
		{
			name:        "update gauge",
			path:        fmt.Sprintf("%s/update/", httpTestServer.URL),
			body:        validGagugeJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusOK,
				body: validGaugeJSONResp,
			},
		},
		{
			name:        "update unknown metric kind",
			path:        fmt.Sprintf("%s/update/", httpTestServer.URL),
			body:        notValidCounterJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "update empty id metric",
			path:        fmt.Sprintf("%s/update/", httpTestServer.URL),
			body:        notValidGaugeJSON,
			contentType: "application/json",
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "update metrics with wrong content type",
			path:        fmt.Sprintf("%s/update/", httpTestServer.URL),
			body:        validGagugeJSON,
			contentType: "text/plain",
			expected: result{
				code: http.StatusUnsupportedMediaType,
			},
		},
		{
			name:        "update metrics with invalid Json",
			path:        fmt.Sprintf("%s/update/", httpTestServer.URL),
			body:        invalidJson,
			contentType: "application/json",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		assert := assert.New(t)
		t.Run(tc.name, func(t *testing.T) {
			tr := &http.Transport{}
			client := &http.Client{Transport: tr}
			res, err := client.Post(tc.path, tc.contentType, strings.NewReader(tc.body))
			require.NoError(t, err)
			assert.Equal(tc.expected.code, res.StatusCode)
			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				assert.NotEmpty(string(respBody))
				assert.Equal(strings.TrimSpace(tc.expected.body), strings.TrimSpace(string(respBody)))

				defer res.Body.Close()

				log.Println("Request body: ", tc.body)
				log.Println("Expected body: ", tc.expected.body)
				log.Println("Response body: ", string(respBody))
			}
		})
	}
}

type MockStorage struct {
	storage.Storage
	getAllGaugesFunc   func() ([]storage.GaugeMetric, error)
	getAllCountersFunc func() ([]storage.CounterMetric, error)
}

func (m *MockStorage) GetAllGauges() ([]storage.GaugeMetric, error) {
	if m.getAllGaugesFunc != nil {
		return m.getAllGaugesFunc()
	}
	return nil, errors.New("not implemented")
}

func (m *MockStorage) GetAllCounters() ([]storage.CounterMetric, error) {
	if m.getAllCountersFunc != nil {
		return m.getAllCountersFunc()
	}
	return nil, errors.New("not implemented")
}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		name               string
		getAllGaugesFunc   func() ([]storage.GaugeMetric, error)
		getAllCountersFunc func() ([]storage.CounterMetric, error)
		expectedCode       int
		expectedBody       string
	}{
		{
			name: "Success",
			getAllGaugesFunc: func() ([]storage.GaugeMetric, error) {
				return []storage.GaugeMetric{
					{Name: "gauge1", Value: 1.23},
					{Name: "gauge2", Value: 4.56},
				}, nil
			},
			getAllCountersFunc: func() ([]storage.CounterMetric, error) {
				return []storage.CounterMetric{
					{Name: "counter1", Value: 10},
					{Name: "counter2", Value: 20},
				}, nil
			},
			expectedCode: http.StatusOK,
			expectedBody: "<html><body>" +
				"<h2>Gauge metrics:</h2><ul>" +
				"<li>gauge1 = 1.23</li>" +
				"<li>gauge2 = 4.56</li>" +
				"</ul>" +
				"<h2>Counter metrics:</h2><ul>" +
				"<li>counter1 = 10</li>" +
				"<li>counter2 = 20</li>" +
				"</ul></body></html>",
		},
		{
			name: "ErrorGetAllGauges",
			getAllGaugesFunc: func() ([]storage.GaugeMetric, error) {
				return nil, errors.New("error fetching gauges")
			},
			getAllCountersFunc: func() ([]storage.CounterMetric, error) {
				return []storage.CounterMetric{
					{Name: "counter1", Value: 10},
					{Name: "counter2", Value: 20},
				}, nil
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
		{
			name: "ErrorGetAllCounters",
			getAllGaugesFunc: func() ([]storage.GaugeMetric, error) {
				return []storage.GaugeMetric{
					{Name: "gauge1", Value: 1.23},
					{Name: "gauge2", Value: 4.56},
				}, nil
			},
			getAllCountersFunc: func() ([]storage.CounterMetric, error) {
				return nil, errors.New("error fetching counters")
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			// Создаем моковое хранилище
			mockStorage := &MockStorage{
				getAllGaugesFunc:   tt.getAllGaugesFunc,
				getAllCountersFunc: tt.getAllCountersFunc,
			}
			handler := GetAllMetrics(mockStorage)

			// Создаем новый HTTP-запрос
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			if assert.NoError(t, handler(ctx)) {
				assert.Equal(t, tt.expectedCode, rec.Code)

				if tt.expectedCode == http.StatusOK {
					assert.Equal(t, "text/html", rec.Header().Get("Content-Type"))
				}

				assert.Equal(t, strings.TrimSpace(tt.expectedBody), strings.TrimSpace(rec.Body.String()))

				log.Println("Request body: ", tt.expectedBody)
				log.Println("Response body: ", rec.Body.String())
			}
		})
	}
}
