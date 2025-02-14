package export

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"

	"github.com/Sofja96/go-metrics.git/internal/agent/gzip"
	"github.com/Sofja96/go-metrics.git/internal/agent/hash"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/proto"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

type GRPCClient struct {
	Client proto.MetricsClient
	conn   *grpc.ClientConn
}

// NewGRPCClient creates a new gRPC client.
func NewGRPCClient(addr string) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewMetricsClient(conn)
	return &GRPCClient{
		Client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection.
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func addMetadata(ctx context.Context, data []byte, key string) (context.Context, error) {
	md := metadata.New(map[string]string{
		"content-encoding": "gzip",
	})

	realIP, err := utils.GetLocalIP()
	if err != nil {
		return nil, fmt.Errorf("error getting local IP: %w", err)
	}
	md.Set("X-Real-IP", realIP)

	if len(key) != 0 {
		hmac, err := hash.ComputeHmac256([]byte(key), data)
		if err != nil {
			return nil, fmt.Errorf("error computing HMAC: %w", err)
		}
		md.Set("HashSHA256", hmac)
	}

	return metadata.NewOutgoingContext(ctx, md), nil
}

// UpdateMetrics sends metrics to the gRPC server.
func (c *GRPCClient) UpdateMetrics(ctx context.Context, metrics []*proto.Metric, post models.PostRequest) (*proto.UpdateMetricsResponse, error) {
	var dataToSend []byte

	compressedMetrics, err := gzip.Compress(metrics)
	if err != nil {
		return nil, fmt.Errorf("compression error %v", err)

	}

	dataToSend = compressedMetrics
	if post.PublicKey != nil {
		encryptedData, err := EncryptWithPublicKey(compressedMetrics, post.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("error encrypting data: %w", err)
		}
		dataToSend = encryptedData
	}

	ctx, err = addMetadata(ctx, dataToSend, post.Key)
	if err != nil {
		return nil, fmt.Errorf("error adding metadata: %w", err)
	}

	req := &proto.UpdateMetricsRequest{
		CompressedData: dataToSend,
	}

	res, err := c.Client.UpdateMetrics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error sending metrics via gRPC: %w", err)
	}

	if !res.GetSuccess() {
		log.Println("Server returned an error while updating metrics")
	}

	return res, nil
}
