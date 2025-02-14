package grpcserver

import (
	"context"
	"go.uber.org/zap"

	"github.com/Sofja96/go-metrics.git/internal/proto"
	"github.com/Sofja96/go-metrics.git/internal/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HMACInterceptor - интерцептор для проверки HMAC-подписи данных.
func HMACInterceptor(logger *zap.SugaredLogger, key []byte) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
		}

		clientHmac := md.Get("HashSHA256")
		if len(clientHmac) == 0 {
			return handler(ctx, req)
		}

		if len(key) == 0 {
			return handler(ctx, req)
		}

		updateReq, ok := req.(*proto.UpdateMetricsRequest)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "invalid request type")
		}

		serverHmac := utils.ComputeHmac256(key, updateReq.CompressedData)

		if clientHmac[0] != serverHmac {
			return nil, status.Errorf(codes.Unauthenticated, "HMAC verification failed")
		}

		logger.Infof("Hash is equal. Requests is successfully")

		return handler(ctx, req)
	}
}
