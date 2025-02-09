package grpcserver

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Sofja96/go-metrics.git/internal/proto"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

func TestHMACInterceptor(t *testing.T) {
	logger := zap.NewNop().Sugar()
	key := []byte("test-key")
	interceptor := HMACInterceptor(logger, key)

	req := &proto.UpdateMetricsRequest{
		CompressedData: []byte("test-data"),
	}

	hmac := utils.ComputeHmac256(key, req.CompressedData)

	t.Run("ValidHMAC", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("HashSHA256", hmac))

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if err != nil {
			t.Fatalf("Interceptor returned an error: %v", err)
		}
	})
	t.Run("NoHMAC", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if err != nil {
			t.Fatalf("Interceptor returned an error: %v", err)
		}
	})
	t.Run("InvalidHMAC", func(t *testing.T) {
		invalidHMAC := "invalid-hmac"
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("HashSHA256", invalidHMAC))

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if status.Code(err) != codes.Unauthenticated {
			t.Fatalf("Expected Unauthenticated error, got %v", err)
		}
	})
	t.Run("MissingMetadata", func(t *testing.T) {
		ctx := context.Background()

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("Expected InvalidArgument error, got %v", err)
		}
	})
	t.Run("InvalidRequestType", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("HashSHA256", hmac))

		invalidReq := "invalid-request-type"

		_, err := interceptor(ctx, invalidReq, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("Expected InvalidArgument error, got %v", err)
		}
	})
	t.Run("EmptyKey", func(t *testing.T) {
		emptyKeyInterceptor := HMACInterceptor(logger, []byte{}) // Создаем интерцептор с пустым ключом

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("HashSHA256", hmac))

		_, err := emptyKeyInterceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if err != nil {
			t.Fatalf("Interceptor returned an error with empty key: %v", err)
		}
	})
}
