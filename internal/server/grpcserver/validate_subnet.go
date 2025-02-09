package grpcserver

import (
	"context"
	"net"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ValidateTrustedSubnetInterceptor - интерцептор для проверки доверенной подсети по заголовку X-Real-IP.
func ValidateTrustedSubnetInterceptor(trustedSubnet string, logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet == "" {
			logger.Info("No trusted subnet configured, skipping validation.")
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("Missing metadata")
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		realIPs := md.Get("x-real-ip")
		if len(realIPs) == 0 {
			logger.Warn("Missing X-Real-IP header")
			return nil, status.Errorf(codes.Unauthenticated, "missing X-Real-IP header")
		}
		realIP := strings.TrimSpace(realIPs[0])

		_, cidr, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Error("Invalid trusted subnet configuration", "error", err)
			return nil, status.Errorf(codes.Internal, "invalid trusted subnet configuration")
		}

		ip := net.ParseIP(realIP)
		if ip == nil {
			logger.Warn("Invalid IP address in X-Real-IP header", "ip", realIP)
			return nil, status.Errorf(codes.Unauthenticated, "invalid IP address in X-Real-IP header")
		}

		if !cidr.Contains(ip) {
			logger.Warn("Access denied: IP not in trusted subnet",
				"ip ", ip.String(),
				" trustedSubnet ", trustedSubnet)
			return nil, status.Errorf(codes.PermissionDenied, "access denied: IP not in trusted subnet")
		}

		logger.Info("Access granted: IP in trusted subnet",
			"ip ", ip.String(),
			" trustedSubnet ", trustedSubnet)

		return handler(ctx, req)
	}
}
