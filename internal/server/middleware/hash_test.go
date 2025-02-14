package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/Sofja96/go-metrics.git/internal/utils"
)

func TestHashMacMiddleware(t *testing.T) {
	key := []byte("test-secret-key")
	e := echo.New()

	e.Use(HashMacMiddleware(key))

	e.POST("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	t.Run("valid hash", func(t *testing.T) {
		body := []byte("test body")
		hash := utils.ComputeHmac256(key, body)

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
		req.Header.Set("Hashsha256", hash)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		responseHash := rec.Header().Get("HashSHA256")
		if responseHash == "" {
			t.Error("HashSHA256 header missing in response")
		}
	})

	t.Run("invalid hash", func(t *testing.T) {
		body := []byte("test body")
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
		req.Header.Set("Hashsha256", "invalid-hash")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}
