package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/Sofja96/go-metrics.git/internal/models"
	middleware2 "github.com/Sofja96/go-metrics.git/internal/server/middleware"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	storagemock "github.com/Sofja96/go-metrics.git/internal/server/storage/mocks"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

type mocks struct {
	storage *storagemock.MockStorage
	logger  zap.SugaredLogger
}

func TestWebhook(t *testing.T) {
	type (
		args struct {
			metricsName         string
			metricsValueCounter int64
			metricValueGauge    float64
			invalidValue        string
		}
		mockBehavior func(m *mocks, args args)
	)

	tests := []struct {
		name               string
		path               string
		args               args
		mockBehavior       mockBehavior
		expectedStatusCode int
		method             string
		expectedResponse   string // Мы добавляем это для проверки ответа
	}{
		{
			name:   "PushCounterSuccess",
			path:   "/update/counter/PollCount/10",
			method: http.MethodPost,
			args: args{
				metricsName:         "PollCount",
				metricsValueCounter: 10,
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().UpdateCounter(gomock.Any(), args.metricsName, args.metricsValueCounter).Return(int64(10), nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:   "PushGaugeSuccess",
			path:   "/update/gauge/Alloc/13.123",
			method: http.MethodPost,
			args: args{
				metricsName:      "Alloc",
				metricValueGauge: 13.123,
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().UpdateGauge(gomock.Any(), args.metricsName, args.metricValueGauge).Return(13.123, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "PushUnknownMetricKind",
			path:               "/update/unknown/Alloc/12.123",
			method:             http.MethodPost,
			mockBehavior:       func(m *mocks, args args) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid metric type. Metric type can only be 'gauge' or 'counter'",
		},
		{
			name:               "PushWithoutNameMetric",
			path:               "/update/Alloc/12.123",
			method:             http.MethodPost,
			mockBehavior:       func(m *mocks, args args) {},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   "{\"message\":\"Not Found\"}\n",
		},
		{
			name:   "PushCounterWithInvalidValueForTypeCounter",
			path:   "/update/counter/Alloc/18446744073709551617",
			method: http.MethodPost,
			args: args{
				invalidValue: "18446744073709551617",
			},
			mockBehavior: func(m *mocks, args args) {
				_, err := strconv.ParseInt(args.invalidValue, 10, 64)
				assert.Error(t, err)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "incorrect values(int) of metric: 18446744073709551617",
		},
		{
			name:   "PushCounterWithInvalidValue",
			path:   "/update/gauge/Alloc/10\\.0",
			method: http.MethodPost,
			args: args{
				invalidValue: "10\\.0",
			},
			mockBehavior: func(m *mocks, args args) {
				_, err := strconv.ParseInt(args.invalidValue, 10, 64)
				assert.Error(t, err)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "incorrect values(float) of metric: 10\\.0",
		},
		{
			name:   "PushCounterWithErrorUpdateCounter",
			path:   "/update/counter/PollCount/10",
			method: http.MethodPost,
			args: args{
				metricsName:         "PollCount",
				metricsValueCounter: 10,
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().UpdateCounter(gomock.Any(), args.metricsName, args.metricsValueCounter).Return(int64(10), fmt.Errorf("error insert counter"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "{\"message\":\"Internal Server Error\"}\n",
		},
		{
			name:   "PushCounterWithErrorUpdateGauge",
			path:   "/update/gauge/Alloc/13.123",
			method: http.MethodPost,
			args: args{
				metricsName:      "Alloc",
				metricValueGauge: 13.123,
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().UpdateGauge(gomock.Any(), args.metricsName, args.metricValueGauge).Return(13.123, fmt.Errorf("error insert gauge"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "{\"message\":\"Internal Server Error\"}\n",
		},
		{
			name:               "PushMethodGetError",
			method:             http.MethodGet,
			path:               "/update/gauge/Alloc/13.123",
			mockBehavior:       func(m *mocks, args args) {},
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedResponse:   "{\"message\":\"Method Not Allowed\"}\n", // Ожидаемый текст
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m, tt.args)
			e := echo.New()
			e.Use(middleware2.WithLogging(m.logger))
			e.POST("/update/:typeM/:nameM/:valueM", Webhook(m.storage))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.path, nil)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			assert.Equal(t, tt.expectedResponse, w.Body.String())

		})
	}
}

func TestValueMetric(t *testing.T) {
	type (
		args struct {
			metricsName string
		}
		mockBehavior func(m *mocks, args args)
	)

	tests := []struct {
		name               string
		path               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedBody       string
		args               args
		method             string
	}{
		{
			name: "getCounterSuccess",
			path: "/value/counter/PollCount1",
			args: args{
				metricsName: "PollCount1",
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetCounterValue(gomock.Any(), args.metricsName).Return(int64(10), true)
			},
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "10",
		},
		{
			name: "GetGaugeSuccess",
			path: "/value/gauge/Alloc1",
			args: args{
				metricsName: "Alloc1",
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetGaugeValue(gomock.Any(), args.metricsName).Return(10.25, true)
			},
			method:             http.MethodGet,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "10.25",
		},
		{
			name:               "GetUnknownMetricKind",
			path:               "/value/unknown/Alloc",
			mockBehavior:       func(m *mocks, args args) {},
			expectedStatusCode: http.StatusNotFound,
			expectedBody: "Metric not fount or invalid metric type. " +
				"Metric type can only be 'gauge' or 'counter'",
		},
		{
			name: "GetUnknownCounter",
			path: "/value/counter/unknown",
			args: args{
				metricsName: "unknown",
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetCounterValue(gomock.Any(), args.metricsName).Return(int64(0), false)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "",
		},
		{
			name: "GetUnknownGauge",
			path: "/value/gauge/unknown",
			args: args{
				metricsName: "unknown",
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetGaugeValue(gomock.Any(), args.metricsName).Return(float64(0), false)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m, tt.args)
			e := echo.New()
			e.Use(middleware2.WithLogging(m.logger))
			e.GET("/value/:typeM/:nameM", ValueMetric(m.storage))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.path, nil)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	type (
		args struct {
			getAllGauges   []storage.GaugeMetric
			getAllCounters []storage.CounterMetric
		}
		mockBehavior func(m *mocks, args args)
	)

	tests := []struct {
		name                 string
		args                 args
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "SuccessGetAllMetrics",
			args: args{
				getAllGauges: []storage.GaugeMetric{
					{Name: "gauge1", Value: 1.23},
					{Name: "gauge2", Value: 4.56},
				},
				getAllCounters: []storage.CounterMetric{
					{Name: "counter1", Value: 10},
					{Name: "counter2", Value: 20},
				},
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetAllGauges(gomock.Any()).Return([]storage.GaugeMetric{
					{Name: "gauge1", Value: 1.23},
					{Name: "gauge2", Value: 4.56},
				}, nil)
				m.storage.EXPECT().GetAllCounters(gomock.Any()).Return([]storage.CounterMetric{
					{Name: "counter1", Value: 10},
					{Name: "counter2", Value: 20},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponseBody: "<html><body>" +
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
			args: args{
				getAllGauges: []storage.GaugeMetric{},
				getAllCounters: []storage.CounterMetric{
					{Name: "counter1", Value: 10},
					{Name: "counter2", Value: 20},
				},
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetAllGauges(gomock.Any()).Return(nil, errors.New("error get all gauges metrics"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "",
		},
		{
			name: "ErrorGetAllCounters",
			args: args{
				getAllGauges: []storage.GaugeMetric{
					{Name: "gauge1", Value: 1.23},
					{Name: "gauge2", Value: 4.56},
				},
				getAllCounters: []storage.CounterMetric{},
			},
			mockBehavior: func(m *mocks, args args) {
				m.storage.EXPECT().GetAllGauges(gomock.Any()).Return([]storage.GaugeMetric{
					{Name: "gauge1", Value: 1.23},
					{Name: "gauge2", Value: 4.56},
				}, nil)
				m.storage.EXPECT().GetAllCounters(gomock.Any()).Return(nil, errors.New("error get counters gauges metrics"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m, tt.args)
			e := echo.New()
			e.Use(middleware2.WithLogging(m.logger))
			e.GET("/", GetAllMetrics(m.storage))

			w := httptest.NewRecorder()

			r := httptest.NewRequest("GET", "/", nil)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.expectedStatusCode == http.StatusOK {
				assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
			}
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestValueJSON(t *testing.T) {
	type (
		mockBehavior func(m *mocks, args models.Metrics)
	)
	tests := []struct {
		name                 string
		reqBodyFile          string
		expectedResponseBody string
		mockBehavior         mockBehavior
		args                 models.Metrics
		expectedStatusCode   int
		contentType          string
	}{
		{
			name:        "getCounterSuccess",
			reqBodyFile: "./mocks/requests/get_counter_ok.json",
			args: models.Metrics{
				ID:    "PollCount1",
				MType: "counter",
				Delta: utils.IntPtr(2),
				Value: nil,
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().GetCounterValue(gomock.Any(), args.ID).Return(int64(2), true)
			},
			expectedResponseBody: strings.NewReplacer("\n", "", " ", "").
				Replace(utils.GetDataFromFile("./mocks/responses/get_counter_ok.json").String()),
			expectedStatusCode: http.StatusOK,
			contentType:        "application/json",
		},
		{
			name:        "GetGaugeSuccess",
			reqBodyFile: "./mocks/requests/get_gauge_ok.json",
			args: models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Delta: nil,
				Value: utils.FloatPtr(13.175),
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().GetGaugeValue(gomock.Any(), args.ID).Return(13.175, true)
			},
			expectedResponseBody: strings.NewReplacer("\n", "", " ", "").
				Replace(utils.GetDataFromFile("./mocks/responses/get_gauge_ok.json").String()),
			expectedStatusCode: http.StatusOK,
			contentType:        "application/json",
		},
		{
			name:               "GetUnknownMetricKind",
			reqBodyFile:        "./mocks/requests/get_unknown_metric_type.json",
			mockBehavior:       func(m *mocks, args models.Metrics) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponseBody: "Metric not found or invalid metric type. " +
				"Metric type can only be 'gauge' or 'counter'",
			contentType: "application/json",
		},
		{
			name:                 "GetEmptyIdMetric",
			reqBodyFile:          "./mocks/requests/get_empty_id_metric.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "No id metric for counter",
			contentType:          "application/json",
		},
		{
			name:                 "GetMetricsWithWrongContentType",
			reqBodyFile:          "./mocks/requests/get_gauge_ok.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			contentType:          "text/plain",
			expectedStatusCode:   http.StatusUnsupportedMediaType,
			expectedResponseBody: "",
		},
		{
			name:                 "GetMetricsWithInvalidJson",
			reqBodyFile:          "./mocks/requests/get_metrics_invalid_json.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			contentType:          "application/json",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Error in JSON decode: invalid character 'g' looking for beginning of value",
		},
		{
			name:        "GetNotExistsMetricsCounter",
			reqBodyFile: "./mocks/requests/get_counter_error.json",
			args: models.Metrics{
				ID:    "PollCount1",
				MType: "counter",
				Delta: nil,
				Value: nil,
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().GetCounterValue(gomock.Any(), args.ID).Return(int64(0), false)
			},
			contentType:          "application/json",
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "",
		},
		{
			name:        "GetNotExistsMetricsGauge",
			reqBodyFile: "./mocks/requests/get_gauge_error.json",
			args: models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Delta: nil,
				Value: nil,
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().GetGaugeValue(gomock.Any(), args.ID).Return(float64(0), false)
			},
			contentType:          "application/json",
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m, tt.args)
			e := echo.New()
			e.Use(middleware2.WithLogging(m.logger))
			e.POST("/value/", ValueJSON(m.storage))

			w := httptest.NewRecorder()

			r := httptest.NewRequest("POST", "/value/", utils.GetDataFromFile(tt.reqBodyFile))
			r.Header.Set("Content-Type", tt.contentType)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestUpdateJSON(t *testing.T) {
	type (
		mockBehavior func(m *mocks, args models.Metrics)
	)
	tests := []struct {
		name                 string
		reqBodyFile          string
		expectedResponseBody string
		mockBehavior         mockBehavior
		args                 models.Metrics
		expectedStatusCode   int
		contentType          string
	}{
		{
			name:        "updateCounterSuccess",
			reqBodyFile: "./mocks/requests/update_counter_json_ok.json",
			args: models.Metrics{
				ID:    "PollCount1",
				MType: "counter",
				Delta: utils.IntPtr(2),
				Value: nil,
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().UpdateCounter(gomock.Any(), args.ID, *args.Delta).Return(int64(2), nil)
			},
			expectedResponseBody: strings.NewReplacer("\n", "", " ", "").
				Replace(utils.GetDataFromFile("./mocks/responses/update_counter_json_ok.json").String()),
			expectedStatusCode: http.StatusOK,
			contentType:        "application/json",
		},
		{
			name:        "updateCounterError",
			reqBodyFile: "./mocks/requests/update_counter_json_ok.json",
			args: models.Metrics{
				ID:    "PollCount1",
				MType: "counter",
				Delta: utils.IntPtr(2),
				Value: nil,
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().UpdateCounter(gomock.Any(), args.ID, *args.Delta).Return(int64(0), errors.New("error update counter value"))
			},
			expectedResponseBody: "{\"message\":\"Internal Server Error\"}",
			expectedStatusCode:   http.StatusInternalServerError,
			contentType:          "application/json",
		},
		{
			name:        "updateGaugeSuccess",
			reqBodyFile: "./mocks/requests/update_gauge_json_ok.json",
			args: models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Delta: nil,
				Value: utils.FloatPtr(13.175),
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().UpdateGauge(gomock.Any(), args.ID, *args.Value).Return(13.175, nil)
			},
			expectedResponseBody: strings.NewReplacer("\n", "", " ", "").
				Replace(utils.GetDataFromFile("./mocks/responses/update_gauge_json_ok.json").String()),
			expectedStatusCode: http.StatusOK,
			contentType:        "application/json",
		},
		{
			name:        "updateGaugeError",
			reqBodyFile: "./mocks/requests/update_gauge_json_ok.json",
			args: models.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Delta: nil,
				Value: utils.FloatPtr(13.175),
			},
			mockBehavior: func(m *mocks, args models.Metrics) {
				m.storage.EXPECT().UpdateGauge(gomock.Any(), args.ID, *args.Value).Return(float64(0), errors.New("error update gauge value"))
			},
			expectedResponseBody: "{\"message\":\"Internal Server Error\"}",
			expectedStatusCode:   http.StatusInternalServerError,
			contentType:          "application/json",
		},
		{
			name:                 "updateUnknownMetricKind",
			reqBodyFile:          "./mocks/requests/get_unknown_metric_type.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "Invalid metric type. Can only be 'gauge' or 'counter'",
			contentType:          "application/json",
		},
		{
			name:                 "UpdateEmptyIdMetric",
			reqBodyFile:          "./mocks/requests/get_empty_id_metric.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: "No id metric for counter",
			contentType:          "application/json",
		},
		{
			name:                 "UpdateMetricsWithWrongContentType",
			reqBodyFile:          "./mocks/requests/update_gauge_json_ok.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			contentType:          "text/plain",
			expectedStatusCode:   http.StatusUnsupportedMediaType,
			expectedResponseBody: "",
		},
		{
			name:                 "UpdateMetricsWithInvalidJson",
			reqBodyFile:          "./mocks/requests/get_metrics_invalid_json.json",
			mockBehavior:         func(m *mocks, args models.Metrics) {},
			contentType:          "application/json",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Error in JSON decode: invalid character 'g' looking for beginning of value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m, tt.args)
			e := echo.New()
			e.Use(middleware2.WithLogging(m.logger))
			e.POST("/update/", UpdateJSON(m.storage))

			w := httptest.NewRecorder()

			r := httptest.NewRequest("POST", "/update/", utils.GetDataFromFile(tt.reqBodyFile))
			r.Header.Set("Content-Type", tt.contentType)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestUpdatesBatch(t *testing.T) {
	type (
		mockBehavior func(m *mocks, args []models.Metrics)
	)
	tests := []struct {
		name                 string
		reqBodyFile          string
		expectedResponseBody string
		mockBehavior         mockBehavior
		args                 []models.Metrics
		expectedStatusCode   int
		contentType          string
	}{
		{
			name:        "BatchUpdateSuccess",
			reqBodyFile: "./mocks/requests/update_batch_ok.json",
			args: []models.Metrics{
				{MType: "gauge", ID: "Alloc", Value: utils.FloatPtr(1.98)},
				{MType: "counter", ID: "PollCount1", Delta: utils.IntPtr(2)},
			},
			mockBehavior: func(m *mocks, args []models.Metrics) {
				m.storage.EXPECT().BatchUpdate(gomock.Any(), args).Return(nil)
			},
			expectedResponseBody: strings.NewReplacer("\n", "", " ", "").
				Replace(utils.GetDataFromFile("./mocks/responses/update_batch_ok.json").String()),
			expectedStatusCode: http.StatusOK,
			contentType:        "application/json",
		},
		{
			name:        "BatchUpdateError",
			reqBodyFile: "./mocks/requests/update_batch_ok.json",
			args: []models.Metrics{
				{MType: "gauge", ID: "Alloc", Value: utils.FloatPtr(1.98)},
				{MType: "counter", ID: "PollCount1", Delta: utils.IntPtr(2)},
			},
			mockBehavior: func(m *mocks, args []models.Metrics) {
				m.storage.EXPECT().BatchUpdate(gomock.Any(), args).Return(fmt.Errorf("error batch update"))
			},
			expectedResponseBody: "error batch update",
			expectedStatusCode:   http.StatusInternalServerError,
			contentType:          "application/json",
		},
		{
			name:                 "EmptyMetricsArray",
			reqBodyFile:          "./mocks/requests/update_batch_empty_metrics.json",
			args:                 []models.Metrics{},
			mockBehavior:         func(m *mocks, args []models.Metrics) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "metrics is empty",
			contentType:          "application/json",
		},
		{
			name:               "InvalidJSONFormat",
			reqBodyFile:        "./mocks/requests/update_batch_invalid_json.json",
			mockBehavior:       func(m *mocks, args []models.Metrics) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponseBody: "\"invalid json code=400, message=Syntax error: offset=38, " +
				"error=invalid character 'A' looking for beginning of value, " +
				"internal=invalid character 'A' looking for beginning of value\"",
			contentType: "application/json",
		},
		{
			name:                 "UpdateMetricsWithWrongContentType",
			reqBodyFile:          "./mocks/requests/update_batch_ok.json",
			mockBehavior:         func(m *mocks, args []models.Metrics) {},
			contentType:          "text/plain",
			expectedStatusCode:   http.StatusUnsupportedMediaType,
			expectedResponseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m, tt.args)
			e := echo.New()
			e.Use(middleware2.GzipMiddleware())
			e.Use(middleware2.WithLogging(m.logger))
			e.POST("/updates/", UpdatesBatch(m.storage))

			w := httptest.NewRecorder()

			r := httptest.NewRequest("POST", "/updates/", utils.GetDataFromFile(tt.reqBodyFile))
			r.Header.Set("Content-Type", tt.contentType)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, strings.TrimSpace(w.Body.String()))

		})
	}
}

func TestPing(t *testing.T) {
	type (
		mockBehavior func(m *mocks)
	)

	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "PingSuccess",
			mockBehavior: func(m *mocks) {
				m.storage.EXPECT().Ping(gomock.Any()).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Connection database is OK",
		},
		{
			name: "PingError",
			mockBehavior: func(m *mocks) {
				m.storage.EXPECT().Ping(gomock.Any()).Return(errors.New("error execute ping"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Connection database is NOT ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			m := &mocks{
				storage: storagemock.NewMockStorage(c),
				logger:  *zap.NewNop().Sugar(),
			}

			tt.mockBehavior(m)
			e := echo.New()
			e.Use(middleware2.WithLogging(m.logger))
			e.GET("/ping", Ping(m.storage))

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/ping", nil)

			e.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())

		})
	}
}
