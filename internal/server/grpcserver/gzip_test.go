package grpcserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Sofja96/go-metrics.git/internal/proto"
)

func TestGzipInterceptor(t *testing.T) {
	logger := zap.NewNop().Sugar()
	interceptor := GzipInterceptor(logger)

	t.Run("SuccessfulDecompression", func(t *testing.T) {
		metrics := []*proto.Metric{
			{Id: "test1", Type: "gauge", Value: 1.23},
			{Id: "test2", Type: "counter", Delta: 42},
		}

		jsonData, err := json.Marshal(metrics)
		if err != nil {
			t.Fatalf("Failed to marshal metrics: %v", err)
		}

		var compressedData bytes.Buffer
		gz := gzip.NewWriter(&compressedData)
		if _, err := gz.Write(jsonData); err != nil {
			t.Fatalf("Failed to compress data: %v", err)
		}
		if err := gz.Close(); err != nil {
			t.Fatalf("Failed to close gzip writer: %v", err)
		}

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("content-encoding", "gzip"))

		req := &proto.UpdateMetricsRequest{
			CompressedData: compressedData.Bytes(),
		}

		_, err = interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			updateReq, ok := req.(*proto.UpdateMetricsRequest)
			if !ok {
				t.Fatal("Request is not of type UpdateMetricsRequest")
			}

			if len(updateReq.Metrics) != 2 {
				t.Fatalf("Expected 2 metrics, got %d", len(updateReq.Metrics))
			}

			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if err != nil {
			t.Fatalf("Interceptor returned an error: %v", err)
		}
	})

	t.Run("MissingMetadata", func(t *testing.T) {
		ctx := context.Background()

		req := &proto.UpdateMetricsRequest{
			CompressedData: []byte{},
		}

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			updateReq, ok := req.(*proto.UpdateMetricsRequest)
			if !ok {
				t.Fatal("Request is not of type UpdateMetricsRequest")
			}

			if len(updateReq.Metrics) != 0 {
				t.Fatalf("Expected no metrics, got %d", len(updateReq.Metrics))
			}

			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if err != nil {
			t.Fatalf("Interceptor returned an error: %v", err)
		}
	})
	t.Run("DecompressionError", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("content-encoding", "gzip"))

		req := &proto.UpdateMetricsRequest{
			CompressedData: []byte{0x1f, 0x8b}, // Невалидные gzip данные
		}

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})

		if err == nil {
			t.Fatal("Expected error but got none")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got %v", err)
		}

		if st.Code() != codes.InvalidArgument {
			t.Fatalf("Expected InvalidArgument error, got %v", st.Code())
		}
	})
	t.Run("JsonUnmarshalError", func(t *testing.T) {
		invalidData := []byte{0x1, 0x2, 0x3}

		var compressedData bytes.Buffer
		gz := gzip.NewWriter(&compressedData)
		if _, err := gz.Write(invalidData); err != nil {
			t.Fatalf("Failed to compress data: %v", err)
		}
		if err := gz.Close(); err != nil {
			t.Fatalf("Failed to close gzip writer: %v", err)
		}

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("content-encoding", "gzip"))

		req := &proto.UpdateMetricsRequest{
			CompressedData: compressedData.Bytes(),
		}

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, nil
		})

		if err == nil {
			t.Fatal("Expected error but got none")
		}

		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("Expected gRPC status error, got %v", err)
		}

		if st.Code() != codes.Internal {
			t.Fatalf("Expected Internal error, got %v", st.Code())
		}
	})
	t.Run("WrongRequestType", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("content-encoding", "gzip"))

		req := &proto.UpdateMetricsResponse{} // Неправильный тип запроса

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			_, ok := req.(*proto.UpdateMetricsRequest)
			if ok {
				t.Fatal("Expected request of type UpdateMetricsResponse, but got UpdateMetricsRequest")
			}
			return &proto.UpdateMetricsResponse{Success: true}, nil
		})
		if err != nil {
			t.Fatalf("Interceptor returned an error: %v", err)
		}
	})
	t.Run("EmptyCompressedData", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("content-encoding", "gzip"))
		req := &proto.UpdateMetricsRequest{
			CompressedData: []byte{}, // Пустые сжатые данные
		}

		_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
			updateReq, ok := req.(*proto.UpdateMetricsRequest)
			if !ok {
				t.Fatal("Request is not of type UpdateMetricsRequest")
			}

			if len(updateReq.Metrics) != 0 {
				t.Fatalf("Expected no metrics, got %d", len(updateReq.Metrics))
			}

			return &proto.UpdateMetricsResponse{Success: true}, nil
		})

		if err != nil {
			t.Fatalf("Interceptor returned an error: %v", err)
		}
	})

}
