package grpcserver

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor - интерцептор для логирования запросов и ответов.
func LoggingInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		logger.Infof("gRPC method %s called", info.FullMethod)

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			logger.Infof("Metadata: %v", md)
		}

		resp, err := handler(ctx, req)

		st, _ := status.FromError(err)
		statusCode := st.Code().String()

		duration := time.Since(start)

		var size int
		if resp != nil {
			jsonResp, _ := json.Marshal(resp)
			size = len(jsonResp)
		}

		if err != nil {
			logger.Errorf("gRPC method %s failed with code %s: %s, "+
				"duration: %s", info.FullMethod, st.Code(), st.Message(), duration)
		}
		logger.Infof("gRPC method %s completed successfully, duration: %s, "+
			"status: %s, size: %d", info.FullMethod, duration, statusCode, size)

		return resp, err
	}
}
