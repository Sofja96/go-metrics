package handlers

import (
	"bytes"
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllMetrics(t *testing.T) {
	s := storage.NewMemStorage()
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
			name: "Push method POST",
			path: fmt.Sprintf("%s/update/gauge/Alloc/13.123", httpTestServer.URL),

			expected: result{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "Push counter",
			path: fmt.Sprintf(httpTestServer.URL),
			expected: result{
				code: http.StatusOK,
				body: "{}",
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
			fmt.Sprintln(res.Body, res.StatusCode)
			fmt.Sprintln(res.StatusCode, http.MethodConnect)
			//	defer res.Body.Close()
			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(res.Body)
				fmt.Sprintln(res.Body, res.StatusCode)
				require.NoError(t, err)
				assert.NotEmpty(len(respBody))
				defer res.Body.Close()
			}
		})
	}
}

func TestWebhook(t *testing.T) {
	s := storage.NewMemStorage()
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
			//res, err := http.Post("http://127.0.0.1:8080"+tc.path, "text/plain", nil)
			require.NoError(t, err)
			assert.Equal(tc.expected.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}

func TestGetMetric(t *testing.T) {
	s := storage.NewMemStorage()
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
				body: "10",
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
	bodyReader := bytes.NewReader([]byte{})
	resp, err := http.Post(fmt.Sprintf("%s/update/counter/PollCount1/10", httpTestServer.URL), "text/plain", bodyReader)
	resp1, err := http.Post(fmt.Sprintf("%s/update/gauge/Alloc1/10", httpTestServer.URL), "text/plain", bodyReader)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	defer resp1.Body.Close()

	fmt.Println("Status:", resp.Status)
	fmt.Println("POST:", resp.Request)
	fmt.Println("Status:", resp1.Status)
	fmt.Println("POST:", resp1.Request)

	for _, tc := range tt {
		assert := assert.New(t)
		t.Run(tc.name, func(t *testing.T) {
			tr := &http.Transport{}
			client := &http.Client{Transport: tr}
			res, err := client.Get(tc.path)
			defer res.Body.Close()
			require.NoError(t, err)
			assert.Equal(tc.expected.code, res.StatusCode)
			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(res.Body)
				assert.NotEmpty(string(respBody))
				assert.Equal(tc.expected.body, string(respBody))
				require.NoError(t, err)
			}
		})
	}
}
