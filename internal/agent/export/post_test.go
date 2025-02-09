package export

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"

	"github.com/Sofja96/go-metrics.git/internal/agent/envs"
	mockproto "github.com/Sofja96/go-metrics.git/internal/agent/export/mocks"
	"github.com/Sofja96/go-metrics.git/internal/agent/gzip"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/proto"
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
			err := PostBatch(tt.client, "http://example.com", data, models.PostRequest{
				Key:       "test-key",
				PublicKey: nil,
			})
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(`{"key": "value"}`)
			_, publicKey := utils.GenerateRsaKeyPair()
			err := PostBatch(tt.client, "http://example.com", data, models.PostRequest{
				Key:       "test-key",
				PublicKey: publicKey,
			})
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

type mocks struct {
	grpcClient *mockproto.MockMetricsClient
}

func TestPostQueries(t *testing.T) {
	type (
		mockBehavior func(m *mocks)
	)

	tests := []struct {
		name         string
		useGRPC      bool
		mockBehavior mockBehavior
		httpClient   *retryablehttp.Client
		channelData  []byte
		expectedLogs []string
	}{
		{
			name:    "successful HTTP request",
			useGRPC: false,
			mockBehavior: func(m *mocks) {
			},
			httpClient: createMockRetryableClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK")),
				}, nil
			}),
			channelData:  []byte(`{"key": "value"}`),
			expectedLogs: []string{""},
		},
		{
			name:    "gRPC client success",
			useGRPC: true,
			mockBehavior: func(m *mocks) {
				m.grpcClient.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&proto.UpdateMetricsResponse{Success: true}, nil)
			},
			channelData: func() []byte {
				metrics := []models.Metrics{
					{
						MType: "gauge",
						ID:    "test_metric",
						Value: utils.FloatPtr(123.45),
					},
				}
				compressedData, err := gzip.Compress(metrics)
				if err != nil {
					t.Fatalf("ошибка сжатия данных: %v", err)
				}
				return compressedData
			}(),
			expectedLogs: []string{"Sending gRPC request with 1 metrics"},
		},
		{
			name:    "gRPC client error",
			useGRPC: true,
			mockBehavior: func(m *mocks) {
				m.grpcClient.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("gRPC error"))
			},
			channelData: func() []byte {
				metrics := []models.Metrics{
					{
						MType: "gauge",
						ID:    "test_metric",
						Value: utils.FloatPtr(123.45),
					},
				}
				compressedData, err := gzip.Compress(metrics)
				if err != nil {
					t.Fatalf("ошибка сжатия данных: %v", err)
				}

				return compressedData
			}(),
			expectedLogs: []string{"Ошибка отправки метрик через gRPC"},
		},
		{
			name:    "gRPC ConvertMetricsError",
			useGRPC: true,
			mockBehavior: func(m *mocks) {
			},
			channelData:  []byte(`{"key": "value"}`),
			expectedLogs: []string{"ошибка преобразования метрик"},
		},
		{
			name:         "channel closed",
			useGRPC:      false,
			mockBehavior: func(m *mocks) {},
			httpClient: createMockRetryableClient(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString("OK")),
				}, nil
			}),
			channelData:  nil,
			expectedLogs: []string{"Канал данных закрыт. Завершаем Worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := &mocks{
				grpcClient: mockproto.NewMockMetricsClient(ctrl),
			}

			tt.mockBehavior(m)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			cfg := &envs.Config{
				UseGRPC:     tt.useGRPC,
				Address:     "example.com",
				GrpcAddress: "localhost:50051",
				HashKey:     "test-key",
			}

			chIn := make(chan []byte, 1)
			if tt.channelData != nil {
				chIn <- tt.channelData
			}
			close(chIn)

			var grpcClient *GRPCClient
			if tt.useGRPC {
				log.Println("grpc иницилизирован")
				grpcClient = &GRPCClient{
					Client: m.grpcClient,
				}
			}

			var logBuffer bytes.Buffer
			log.SetOutput(&logBuffer)
			defer log.SetOutput(os.Stderr)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				PostQueries(ctx, cfg, chIn, nil, grpcClient)
			}()

			wg.Wait()

			time.Sleep(10 * time.Millisecond)

			logOutput := logBuffer.String()
			for _, logMsg := range tt.expectedLogs {
				assert.Contains(t, logOutput, logMsg)
			}

		})
	}
}

func TestLinearBackoff(t *testing.T) {
	retryWaitMin := time.Second * 1
	retryWaitMax := time.Second * 5
	attemptNum := 2

	result := linearBackoff(retryWaitMin, retryWaitMax, attemptNum, nil)
	expected := retryWaitMin + retryWaitMin*time.Duration(2*attemptNum)

	assert.Equal(t, expected, result, "Некорректное время ожидания")
}
