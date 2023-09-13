package handlers

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllMetrics(t *testing.T) {
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
			name: "Push method POST",
			path: "/update/gauge/Alloc/13.123",

			expected: result{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Push counter",
			path: "/",
			expected: result{
				code: http.StatusOK,
				body: "{}",
			},
		},
	}
	for _, tc := range tt {
		e := echo.New()
		s := storage.NewMemStorage()
		e.GET("/", AllMetrics(s))
		e.GET("/value/:typeM/:nameM", ValueMetrics(s))
		e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
		assert := assert.New(t)
		go func() {
			err := e.Start(":8080")
			if err != nil {
				fmt.Print(err)
			}
		}()
		time.Sleep(2 * time.Second)
		t.Run(tc.name, func(t *testing.T) {
			//s := storage.New()
			//AllMetrics(s)
			res, err := http.Get("http://127.0.0.1:8080" + tc.path)
			require.NoError(t, err)
			assert.Equal(tc.expected.code, res.StatusCode)
			//	defer res.Body.Close()
			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.NotEmpty(len(respBody))
				defer res.Body.Close()

			}
		})
	}
}

//func TestWebhook(t *testing.T) {
//	type result struct {
//		code int
//		body string
//	}
//
//	tt := []struct {
//		name     string
//		path     string
//		expected result
//	}{
//		{
//			name: "Push counter",
//			path: "/update/counter/PollCount/10",
//			expected: result{
//				code: http.StatusOK,
//			},
//		},
//		{
//			name: "Push gauge",
//			path: "/update/gauge/Alloc/13.123",
//			expected: result{
//				code: http.StatusOK,
//			},
//		},
//		{
//			name: "Push unknown metric kind",
//			path: "/update/unknown/Alloc/12.123",
//			expected: result{
//				code: http.StatusBadRequest,
//			},
//		},
//		{
//			name: "Push without name metric",
//			path: "/update/Alloc/12.123",
//			expected: result{
//				code: http.StatusNotFound,
//			},
//		},
//		{
//			name: "Push counter with invalid name",
//			path: "/update/counter/Alloc/18446744073709551617",
//			expected: result{
//				code: http.StatusBadRequest,
//			},
//		},
//		{
//			name: "Push counter with invalid value",
//			path: "/update/gauge/PollCount/10\\.0",
//
//			expected: result{
//				code: http.StatusBadRequest,
//			},
//		},
//		{
//			name: "Push method get",
//			path: "/",
//
//			expected: result{
//				code: http.StatusMethodNotAllowed,
//			},
//		},
//	}
//	//e := echo.New()
//	////assert := assert.New(t)
//	////	Webhook(s)
//	//e.Start("localhost:8081")
//	//if err != nil {
//	//	log.Fatal(err)
//	//}
//	for _, tc := range tt {
//		t.Run(tc.name, func(t *testing.T) {
//
//			//	s := storage.New()
//			//e := echo.New()
//			assert := assert.New(t)
//			e := echo.New()
//			////assert := assert.New(t)
//			////	Webhook(s)
//			e.Start("localhost:8081")
//			//	Webhook(s)
//			//err := e.Start("localhost:8081")
//			//if err != nil {
//			//	log.Fatal(err)
//			//}
//			//e.Start("8080")
//			res, err := http.Post("http://127.0.0.1:8081"+tc.path, "text/plain", nil)
//			require.NoError(t, err)
//			assert.Equal(tc.expected.code, res.StatusCode)
//			defer res.Body.Close()
//		})
//	}
//}

//func TestGetMetric(t *testing.T) {
//	type result struct {
//		code int
//		body string
//	}
//
//	tt := []struct {
//		name     string
//		path     string
//		expected result
//	}{
//		{
//			name: "Get counter",
//			path: "/value/counter/PollCount",
//			expected: result{
//				code: http.StatusOK,
//				body: "9601",
//			},
//		},
//		{
//			name: "Get gauge",
//			path: "/value/gauge/Alloc",
//			expected: result{
//				code: http.StatusOK,
//				body: "11.345",
//			},
//		},
//		{
//			name: "Get unknown metric kind",
//			path: "/value/unknown/Alloc",
//			expected: result{
//				code: http.StatusNotFound,
//			},
//		},
//		{
//			name: "Get unknown counter",
//			path: "/value/counter/unknown",
//			expected: result{
//				code: http.StatusNotFound,
//			},
//		},
//		{
//			name: "Get unknown gauge",
//			path: "/value/gauge/unknown",
//			expected: result{
//				code: http.StatusNotFound,
//			},
//		},
//	}
//
//	for _, tc := range tt {
//		//s := storage.NewMemStorage()
//		e := echo.New()
//		assert := assert.New(t)
//		//	ValueMetrics(s)
//		e.Start("8080")
//		t.Run(tc.name, func(t *testing.T) {
//			res, _ := http.Get("http://127.0.0.1:8081" + tc.path)
//			assert.Equal(tc.expected.code, res.StatusCode)
//			defer res.Body.Close()
//			if tc.expected.code == http.StatusOK {
//				respBody, err := io.ReadAll(res.Body)
//				require.NoError(t, err)
//				assert.NotEmpty(string(respBody))
//			}
//		})
//	}
//}

//func testRequest(t *testing.T, ts *httptest.Server,
//	method, path string) (*http.Response, string) {
//	req, err := http.NewRequest(method, ts.URL+path, nil)
//	require.NoError(t, err)
//
//	resp, err := ts.Client().Do(req)
//	require.NoError(t, err)
//	defer resp.Body.Close()
//
//	respBody, err := io.ReadAll(resp.Body)
//	require.NoError(t, err)
//
//	return resp, string(respBody)
//}
//func TestRouter(t *testing.T) {
//	e := echo.New()
//	server := e.Start("localhost")
//	//ts := httptest.NewServer("ValueMetrics")
//	//defer ts.Close()
//
//	tests := []struct {
//		method string
//		url    string
//		want   string
//		status int
//	}{
//		{"POST", "/update/gauge/alloc/4.0", "", 200},
//		{"POST", "/wrongMethod/gauge/alloc/4.0", "404 page not found\n", 404},
//		{"POST", "/update/counter/pollscounter/1", "", 200},
//		{"POST", "/update/counter/pollscounter/f", "", 400},
//		{"GET", "/value/gauge/alloc", "4", 200},
//		{"GET", "/value/counter/pollscounter", "1", 200},
//		{"GET", "/value/counter/wrong", "", 404},
//		{"GET", "/value/wrongtype/wrong", "", 400},
//		{"POST", "/value/counter/pollscounter", "", 405},
//		{"GET", "/", "alloc: 4\npollscounter: 1\n", 200},
//	}
//
//	for _, test := range tests {
//		resp, get := testRequest(t, server, test.method, test.url)
//		defer resp.Body.Close()
//		assert.Equal(t, test.status, resp.StatusCode)
//		assert.Equal(t, test.want, get)
//	}
//}
