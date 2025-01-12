package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/Sofja96/go-metrics.git/internal/utils"
)

func TestLoadPrivatecKey(t *testing.T) {
	privatePath := "private_key.pem"
	publicPath := "public_key.pem"

	defer os.Remove(privatePath)
	defer os.Remove(publicPath)

	privateKey, publicKey := utils.GenerateRsaKeyPair()

	privatePEM := utils.PrivateToString(privateKey)
	publicPEM, _ := utils.PublicToString(publicKey)

	err := utils.ExportToFile(privatePEM, privatePath)
	assert.NoError(t, err)

	err = utils.ExportToFile(publicPEM, publicPath)
	assert.NoError(t, err)

	privateKey, err = LoadPrivateKey("private_key.pem")
	assert.NoError(t, err)
	assert.NotNil(t, privateKey)

	_, err = LoadPrivateKey("nonexistent_file.pem")
	assert.Error(t, err)

	privateKey, err = LoadPrivateKey("")
	assert.NoError(t, err)
	assert.Nil(t, privateKey)
}

func TestMiddlewareDecryption(t *testing.T) {
	privateKey, publicKey := utils.GenerateRsaKeyPair()

	e := echo.New()
	e.Use(DecryptMiddleware(privateKey))

	e.POST("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	t.Run("valid private key", func(t *testing.T) {
		original := []byte("test metrics data")

		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, original, nil)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted)

		decrypted, err := DecryptWithPrivateKey(encrypted, privateKey)
		assert.NoError(t, err)
		assert.NotNil(t, decrypted)
		assert.Equal(t, original, decrypted, "decrypted data should match original")

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(encrypted))
		req.Header.Set("Content-Type", "application/octet-stream")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "expected status 200, got %d", rec.Code)
	})

	t.Run("invalid Content-Type", func(t *testing.T) {
		original := []byte("test metrics data")

		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, original, nil)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(encrypted))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code, "expected status 200 even with invalid content-type")
	})

	t.Run("invalid encrypted data", func(t *testing.T) {
		invalidData := []byte("this is not valid encrypted data")

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(invalidData))
		req.Header.Set("Content-Type", "application/octet-stream")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "error decrypting data", "expected decryption error message")
	})

	t.Run("incorrect private key", func(t *testing.T) {
		invalidPrivateKey, _ := utils.GenerateRsaKeyPair()

		original := []byte("test metrics data")
		encrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, original, nil)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(encrypted))
		req.Header.Set("Content-Type", "application/octet-stream")

		e := echo.New()
		e.Use(DecryptMiddleware(invalidPrivateKey))

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "error decrypting data", "error decrypting chunk")
	})

	t.Run("error reading body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/test", &brokenReader{})
		req.Header.Set("Content-Type", "application/octet-stream")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "error reading request body", "expected error reading request body")
	})

}

// Симуляция ошибки при чтении тела запроса
type brokenReader struct{}

func (r *brokenReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error reading body")
}
