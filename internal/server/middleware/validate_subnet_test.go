package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestValidateTrustedSubnet - тест для проверки middleware.
func TestValidateTrustedSubnet(t *testing.T) {
	tests := []struct {
		name             string
		trustedSubnet    string
		realIP           string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "Valid IP in trusted subnet",
			trustedSubnet:    "192.168.1.0/24",
			realIP:           "192.168.1.100",
			expectedStatus:   http.StatusOK,
			expectedResponse: "OK",
		},
		{
			name:             "IP not in trusted subnet",
			trustedSubnet:    "192.168.1.0/24",
			realIP:           "192.168.2.100",
			expectedStatus:   http.StatusForbidden,
			expectedResponse: "Access denied: IP not in trusted subnet",
		},
		{
			name:             "No X-Real-IP header",
			trustedSubnet:    "192.168.1.0/24",
			realIP:           "",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Missing X-Real-IP header",
		},
		{
			name:             "Invalid trusted subnet CIDR",
			trustedSubnet:    "192.168.1.0/33", // Invalid CIDR
			realIP:           "192.168.1.100",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Invalid trusted subnet configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			e.Use(ValidateTrustedSubnet(tt.trustedSubnet))

			e.GET("/", func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			assert.Equal(t, tt.expectedResponse, rec.Body.String())
		})
	}
}
