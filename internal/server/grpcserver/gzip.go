package grpcserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	model "github.com/Sofja96/go-metrics.git/internal/proto"
)

// GzipInterceptor - интерцептор для обработки gzip-сжатых данных.
func GzipInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("Missing metadata")
			return handler(ctx, req)
		}

		logger.Infof("Incoming metadata: %v", md)

		if encoding := md.Get("content-encoding"); len(encoding) > 0 && encoding[0] == "gzip" {
			logger.Info("Request is compressed with gzip")

			metricsReq, ok := req.(*model.UpdateMetricsRequest)
			if !ok {
				logger.Warn("Request is not UpdateMetricsRequest, skipping decompression")
				return handler(ctx, req)
			}

			if len(metricsReq.CompressedData) == 0 {
				logger.Info("CompressedData is empty, skipping decompression")
				return handler(ctx, req)
			}

			logger.Infof("Compressed request size: %d bytes", len(metricsReq.CompressedData))

			gz, err := gzip.NewReader(bytes.NewReader(metricsReq.CompressedData))
			if err != nil {
				logger.Errorf("Failed to create gzip reader: %v", err)
				return nil, status.Errorf(codes.InvalidArgument, "failed to decompress request: %v", err)
			}
			defer gz.Close()

			decompressedData, err := io.ReadAll(gz)
			if err != nil {
				logger.Errorf("Failed to read decompressed data: %v", err)
				return nil, status.Errorf(codes.InvalidArgument, "failed to read decompressed request: %v", err)
			}
			logger.Infof("Decompressed request size: %d bytes", len(decompressedData))

			var protoMetrics []*model.Metric
			err = json.Unmarshal(decompressedData, &protoMetrics)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed json to model Metrics: %v", err)
			}

			newReq := &model.UpdateMetricsRequest{
				Metrics: protoMetrics,
			}

			req = newReq
		}

		return handler(ctx, req)
	}
}
