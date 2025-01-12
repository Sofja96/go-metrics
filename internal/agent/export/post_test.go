package export

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"

	"github.com/Sofja96/go-metrics.git/internal/utils"
)

func TestPostBatch(t *testing.T) {
	tests := []struct {
		name        string
		client      *retryablehttp.Client
		expectedErr string
	}{
		{
			name: "successful request",
			client: createMockRetryableClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK")),
				}, nil
			}),
			expectedErr: "",
		},
		{
			name: "client error",
			client: createMockRetryableClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewBufferString("Bad Request")),
				}, errors.New("client error")
			}),
			expectedErr: "error connection: POST http://example.com giving up after 1 attempt(s): Post \"http://example.com\": client error",
		},
		{
			name: "server error",
			client: createMockRetryableClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
				}, nil
			}),
			expectedErr: "error connection: POST http://example.com giving up after 1 attempt(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(`{"key": "value"}`)
			err := PostBatch(tt.client, "http://example.com", "test-key", data, nil)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// createMockRetryableClient создаёт mock retryablehttp.Client с кастомным транспортом
func createMockRetryableClient(roundTrip func(req *http.Request) (*http.Response, error)) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.RetryMax = 0 // Отключаем повторы
	client.HTTPClient = &http.Client{
		Transport: roundTripperFunc(roundTrip),
	}
	return client
}

// roundTripperFunc - тип для мокирования RoundTripper
type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestEncryptWithPublicKey(t *testing.T) {
	original := []byte("test metrics data")
	privateKey, publicKey := utils.GenerateRsaKeyPair()

	encrypted, err := EncryptWithPublicKey(original, publicKey)
	assert.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEqual(t, original, encrypted)

	decryptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encrypted, nil)
	assert.NoError(t, err)

	assert.Equal(t, original, decryptedData)
}
