package grpcserver

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestValidateTrustedSubnetInterceptor(t *testing.T) {
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		header         string
		expectedCode   codes.Code
		expectedErrMsg string
	}{
		{
			name:           "No trusted subnet",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.1",
			header:         "x-real-ip",
			expectedCode:   codes.OK,
			expectedErrMsg: "",
		},
		{
			name:           "Missing X-Real-IP",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			header:         "x-real-ip",
			expectedCode:   codes.Unauthenticated,
			expectedErrMsg: "missing metadata",
		},
		{
			name:           "Missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.1",
			header:         "",
			expectedCode:   codes.Unauthenticated,
			expectedErrMsg: "missing X-Real-IP header",
		},
		{
			name:           "IP in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.1",
			header:         "x-real-ip",
			expectedCode:   codes.OK,
			expectedErrMsg: "",
		},
		{
			name:           "IP not in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			header:         "x-real-ip",
			expectedCode:   codes.PermissionDenied,
			expectedErrMsg: "access denied: IP not in trusted subnet",
		},
		{
			name:           "Invalid IP address",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "invalid-ip",
			header:         "x-real-ip",
			expectedCode:   codes.Unauthenticated,
			expectedErrMsg: "invalid IP address in X-Real-IP header",
		},
		{
			name:           "Invalid trusted subnet",
			trustedSubnet:  "invalid-subnet",
			xRealIP:        "192.168.1.1",
			header:         "x-real-ip",
			expectedCode:   codes.Internal,
			expectedErrMsg: "invalid trusted subnet configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := ValidateTrustedSubnetInterceptor(tt.trustedSubnet, logger)

			ctx := context.Background()
			if tt.xRealIP != "" {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(tt.header, tt.xRealIP))
			}

			_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})

			if tt.expectedCode == codes.OK {
				if err != nil {
					t.Fatalf("Expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected an error, got nil")
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("Expected gRPC status error, got %v", err)
				}

				if st.Code() != tt.expectedCode {
					t.Fatalf("Expected error code %v, got %v", tt.expectedCode, st.Code())
				}

				if st.Message() != tt.expectedErrMsg {
					t.Fatalf("Expected error message '%s', got '%s'", tt.expectedErrMsg, st.Message())
				}
			}
		})
	}
}
