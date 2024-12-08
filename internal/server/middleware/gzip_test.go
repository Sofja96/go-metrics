package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware_Compression(t *testing.T) {
	e := echo.New()
	e.Use(GzipMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Gzip!")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res := httptest.NewRecorder()

	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "gzip", res.Header().Get("Content-Encoding"))

	gzReader, err := gzip.NewReader(res.Body)
	assert.NoError(t, err)
	defer gzReader.Close()

	decompressedBody, err := io.ReadAll(gzReader)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, Gzip!", string(decompressedBody))
}

func TestGzipMiddleware_Decompression(t *testing.T) {
	e := echo.New()
	e.Use(GzipMiddleware())

	e.POST("/test", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "")
		}
		return c.String(http.StatusOK, string(body))
	})

	var gzipBody bytes.Buffer
	gzWriter := gzip.NewWriter(&gzipBody)
	_, err := gzWriter.Write([]byte("Hello, Server!"))
	assert.NoError(t, err)
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/test", &gzipBody)
	req.Header.Set("Content-Encoding", "gzip")
	res := httptest.NewRecorder()

	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "Hello, Server!", res.Body.String())
}

func TestGzipMiddleware_NoCompression(t *testing.T) {
	e := echo.New()
	e.Use(GzipMiddleware())

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "No Compression")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	e.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.NotEqual(t, "gzip", res.Header().Get("Content-Encoding"))
	assert.Equal(t, "No Compression", res.Body.String())
}
