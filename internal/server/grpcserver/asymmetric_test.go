package grpcserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sofja96/go-metrics.git/internal/proto"
)

func TestDecryptInterceptor(t *testing.T) {
	logger := zap.NewNop().Sugar()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	interceptor := DecryptInterceptor(logger, privateKey)

	t.Run("Valid decryption", func(t *testing.T) {
		data := []byte("test-data")
		encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, data, nil)
		assert.NoError(t, err)
		assert.NotNil(t, encryptedData)

		req := &proto.UpdateMetricsRequest{CompressedData: encryptedData}

		_, err = interceptor(context.Background(), req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			updateReq, ok := req.(*proto.UpdateMetricsRequest)
			assert.True(t, ok, "Request is not of type UpdateMetricsRequest")
			assert.Equal(t, string(data), string(updateReq.CompressedData), "Decrypted data mismatch")
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})
		assert.NoError(t, err)
	})

	t.Run("Invalid encrypted data", func(t *testing.T) {
		invalidReq := &proto.UpdateMetricsRequest{CompressedData: []byte("invalid-data")}
		_, err = interceptor(context.Background(), invalidReq, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		assert.Error(t, err, "Expected error on invalid encrypted data")
	})

	t.Run("Incorrect private key", func(t *testing.T) {
		incorrectPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		assert.NoError(t, err)

		data := []byte("test-data")
		encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &privateKey.PublicKey, data, nil)
		assert.NoError(t, err)
		assert.NotNil(t, encryptedData)

		incorrectInterceptor := DecryptInterceptor(logger, incorrectPrivateKey)

		req := &proto.UpdateMetricsRequest{CompressedData: encryptedData}

		_, err = incorrectInterceptor(context.Background(), req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			_, ok := req.(*proto.UpdateMetricsRequest)
			assert.True(t, ok, "Request is not of type UpdateMetricsRequest")
			return nil, nil

		})
		assert.Error(t, err, "error decrypting data")
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.Internal, st.Code(), "Expected error code Internal")
	})

	t.Run("Invalid request type", func(t *testing.T) {
		invalidReq := &proto.UpdateMetricsResponse{}
		_, err := interceptor(context.Background(), invalidReq, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})
		assert.Error(t, err, "Expected error on invalid request type")
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.InvalidArgument, st.Code(), "Expected error code InvalidArgument,")
	})
	t.Run("Missing private key", func(t *testing.T) {
		nilInterceptor := DecryptInterceptor(logger, nil) // Создаем интерцептор с nil ключом

		req := &proto.UpdateMetricsRequest{CompressedData: []byte("test-data")}

		_, err := nilInterceptor(context.Background(), req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		assert.NoError(t, err, "Interceptor should not return an error when private key is missing")
	})
}
