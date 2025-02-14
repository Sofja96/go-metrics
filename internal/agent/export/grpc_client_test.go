package export

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	mockproto "github.com/Sofja96/go-metrics.git/internal/agent/export/mocks"
	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/proto"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

func TestNewGRPCClient_MOCK(t *testing.T) {
	addr := "localhost:50051"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockproto.NewMockMetricsClient(ctrl)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	client := &GRPCClient{
		Client: mockClient,
		conn:   conn,
	}

	assert.NotNil(t, client)
	assert.NoError(t, client.Close())
}

func TestNewGRPCClient(t *testing.T) {
	t.Run("NewGRPCClient_SUCCESS", func(t *testing.T) {
		addr := "localhost:50051"

		client, err := NewGRPCClient(addr)
		assert.NoError(t, err)
		assert.NotNil(t, client)

		assert.NoError(t, client.Close())
	})
	// t.Run("NewGRPCClient_ERROR", func(t *testing.T) {
	// 	invalidAddr := ""

	// 	client, err := NewGRPCClient(invalidAddr)
	// 	assert.Error(t, err)
	// 	assert.Contains(t, err.Error(), "failed to connect to gRPC server")
	// 	assert.Nil(t, client)
	// })
}

func TestGRPCClient_UpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockproto.NewMockMetricsClient(ctrl)
	client := &GRPCClient{
		Client: mockClient,
	}

	metrics := []*proto.Metric{
		{Id: "test_metric", Type: "counter", Delta: 1},
	}
	post := models.PostRequest{
		Key: "test_key",
	}
	_, publicKey := utils.GenerateRsaKeyPair()

	t.Run("UpdateMetricsSuccess_WithPublicKey", func(t *testing.T) {
		post := models.PostRequest{
			Key:       "test_key",
			PublicKey: publicKey,
		}

		mockClient.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&proto.UpdateMetricsResponse{Success: true}, nil)

		res, err := client.UpdateMetrics(context.Background(), metrics, post)
		assert.NoError(t, err)
		assert.True(t, res.GetSuccess())
	})
	t.Run("UpdateMetricsSuccess_WithoutPublicKey", func(t *testing.T) {
		mockClient.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&proto.UpdateMetricsResponse{Success: true}, nil)

		res, err := client.UpdateMetrics(context.Background(), metrics, post)
		assert.NoError(t, err)
		assert.True(t, res.GetSuccess())
	})
	t.Run("UpdateMetricsError", func(t *testing.T) {
		post := models.PostRequest{
			Key:       "test_key",
			PublicKey: publicKey,
		}

		mockClient.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc error"))

		res, err := client.UpdateMetrics(context.Background(), metrics, post)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
	t.Run("UpdateMetricsFailureSuccessFlag", func(t *testing.T) {
		mockClient.EXPECT().UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&proto.UpdateMetricsResponse{Success: false}, nil)

		res, err := client.UpdateMetrics(context.Background(), metrics, post)
		assert.NoError(t, err)
		assert.False(t, res.GetSuccess())
	})
}

func TestAddMetadata(t *testing.T) {
	ctx := context.Background()
	data := []byte("test_data")
	key := "test_key"

	t.Run("Test without key", func(t *testing.T) {
		ctx, err := addMetadata(ctx, data, "")
		assert.NoError(t, err)
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, []string{"gzip"}, md.Get("content-encoding"))
	})
	t.Run("Test with key", func(t *testing.T) {
		ctx, err := addMetadata(ctx, data, key)
		assert.NoError(t, err)
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, []string{"gzip"}, md.Get("content-encoding"))
		assert.NotEmpty(t, md.Get("HashSHA256"))
	})
}
