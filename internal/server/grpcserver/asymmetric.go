package grpcserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sofja96/go-metrics.git/internal/proto"
)

// DecryptInterceptor - интерцептор для дешифровки данных, зашифрованных публичным ключом.
func DecryptInterceptor(logger *zap.SugaredLogger, privateKey *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if privateKey == nil {
			logger.Info("Missing privateKey")
			return handler(ctx, req)
		}

		updateReq, ok := req.(*proto.UpdateMetricsRequest)
		if !ok {
			logger.Error("Invalid request type")
			return nil, status.Errorf(codes.InvalidArgument, "invalid request type")
		}

		decryptedData, err := DecryptWithPrivateKey(updateReq.CompressedData, privateKey)
		if err != nil {
			logger.Errorf("Error decrypting data: %v", err)
			return nil, status.Errorf(codes.Internal, "error decrypting data")
		}

		updateReq.CompressedData = decryptedData
		return handler(ctx, updateReq)
	}
}

// DecryptWithPrivateKey - функция для дешифровки данных с использованием приватного ключа.
func DecryptWithPrivateKey(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	chunkSize := privateKey.Size()
	var decryptedData []byte

	for start := 0; start < len(data); start += chunkSize {
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[start:end]

		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("error decrypting chunk: %w", err)
		}

		decryptedData = append(decryptedData, decryptedChunk...)
	}

	return decryptedData, nil
}
