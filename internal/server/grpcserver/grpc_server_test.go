package grpcserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/proto"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/memory"
	storagemock "github.com/Sofja96/go-metrics.git/internal/server/storage/mocks"
	"github.com/Sofja96/go-metrics.git/internal/utils"
)

type mocks struct {
	storage *storagemock.MockStorage
}

func TestUpdateMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := &mocks{
		storage: storagemock.NewMockStorage(ctrl),
	}

	server := &MetricsServer{
		storage: m.storage,
	}

	t.Run("UpdateGauge", func(t *testing.T) {
		m.storage.EXPECT().UpdateGauge(gomock.Any(), "test", 1.23).Return(1.23, nil).Times(1)

		req := &proto.UpdateMetricRequest{
			Metric: &proto.Metric{
				Id:    "test",
				Type:  "gauge",
				Value: 1.23,
			},
		}

		resp, err := server.UpdateMetric(context.Background(), req)
		assert.NoError(t, err, "UpdateMetric returned an error")
		assert.True(t, resp.Success, "expected success response")
	})

	t.Run("UpdateGaugeError", func(t *testing.T) {
		m.storage.EXPECT().UpdateGauge(gomock.Any(), "test", 1.23).Return(float64(0), fmt.Errorf("storage error")).Times(1)

		req := &proto.UpdateMetricRequest{
			Metric: &proto.Metric{
				Id:    "test",
				Type:  "gauge",
				Value: 1.23,
			},
		}

		_, err := server.UpdateMetric(context.Background(), req)
		assert.Error(t, err, "Expected error when updating gauge")
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.Internal, st.Code(), "Expected error code Internal")
	})

	t.Run("UpdateCounter", func(t *testing.T) {
		m.storage.EXPECT().UpdateCounter(gomock.Any(), "test", int64(2)).Return(int64(2), nil).Times(1)

		req := &proto.UpdateMetricRequest{
			Metric: &proto.Metric{
				Id:    "test",
				Type:  "counter",
				Delta: 2,
			},
		}

		resp, err := server.UpdateMetric(context.Background(), req)
		assert.NoError(t, err, "UpdateMetric returned an error")
		assert.True(t, resp.Success, "expected success response")
	})

	t.Run("UpdateCounterError", func(t *testing.T) {
		m.storage.EXPECT().UpdateCounter(gomock.Any(), "test", int64(2)).Return(int64(0), fmt.Errorf("storage error")).Times(1)

		req := &proto.UpdateMetricRequest{
			Metric: &proto.Metric{
				Id:    "test",
				Type:  "counter",
				Delta: 2,
			},
		}

		_, err := server.UpdateMetric(context.Background(), req)
		assert.Error(t, err, "Expected error when updating gauge")
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.Internal, st.Code(), "Expected error code Internal")
	})

	t.Run("UpdateInvalidMetricType", func(t *testing.T) {
		req := &proto.UpdateMetricRequest{
			Metric: &proto.Metric{
				Id:   "invalid",
				Type: "invalid_type",
			},
		}
		_, err := server.UpdateMetric(context.Background(), req)
		assert.Error(t, err, "unsupported metric type")
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.NotFound, st.Code(), "Expected error code NotFound")
	})
}

func TestGetMetric(t *testing.T) {
	store, _ := memory.NewInMemStorage(context.Background(), 300, "", false)
	server := &MetricsServer{
		storage: store,
	}

	reqGauge := &proto.UpdateMetricRequest{
		Metric: &proto.Metric{
			Id:    "testGauge",
			Type:  "gauge",
			Value: 2.71,
		},
	}
	_, err := server.UpdateMetric(context.Background(), reqGauge)
	assert.NoError(t, err, "UpdateMetric returned an error")

	reqCounter := &proto.UpdateMetricRequest{
		Metric: &proto.Metric{
			Id:    "testCounter",
			Type:  "counter",
			Delta: 10,
		},
	}
	_, err = server.UpdateMetric(context.Background(), reqCounter)
	assert.NoError(t, err, "UpdateMetric returned an error")

	srv := &MetricsServer{storage: store}

	t.Run("GetGauge", func(t *testing.T) {
		req := &proto.GetMetricRequest{Name: "testGauge", Type: "gauge"}
		resp, err := srv.GetMetric(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, 2.71, resp.Metric.Value)
	})

	t.Run("GetCounter", func(t *testing.T) {
		req := &proto.GetMetricRequest{Name: "testCounter", Type: "counter"}
		resp, err := srv.GetMetric(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int64(10), resp.Metric.Delta)
	})

	t.Run("GetNonExistentMetricType", func(t *testing.T) {
		req := &proto.GetMetricRequest{Name: "testCounter", Type: "nonexistent"}
		_, err := srv.GetMetric(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.InvalidArgument, st.Code(), "Expected error code InvalidArgument")
	})
	t.Run("GetNonExistentMetricNameGauge", func(t *testing.T) {
		req := &proto.GetMetricRequest{Name: "nonexistent", Type: "gauge"}
		_, err := srv.GetMetric(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.NotFound, st.Code(), "Expected error code NotFound")
	})
	t.Run("GetNonExistentMetricNameCounter", func(t *testing.T) {
		req := &proto.GetMetricRequest{Name: "nonexistent", Type: "counter"}
		_, err := srv.GetMetric(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.NotFound, st.Code(), "Expected error code NotFound")
	})
}

func TestUpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := &mocks{
		storage: storagemock.NewMockStorage(ctrl),
	}

	srv := &MetricsServer{
		storage: m.storage,
	}

	t.Run("BatchUpdateMetrics", func(t *testing.T) {
		m.storage.EXPECT().BatchUpdate(gomock.Any(), []models.Metrics{
			{MType: "gauge", ID: "gauge1", Value: utils.FloatPtr(1.23)},
			{MType: "counter", ID: "counter1", Delta: utils.IntPtr(10)},
		})

		req := &proto.UpdateMetricsRequest{
			Metrics: []*proto.Metric{
				{Id: "gauge1", Type: "gauge", Value: 1.23},
				{Id: "counter1", Type: "counter", Delta: 10},
			},
		}
		resp, err := srv.UpdateMetrics(context.Background(), req)
		assert.NoError(t, err)
		assert.True(t, resp.Success)
	})
	t.Run("BatchUpdateMetricsError", func(t *testing.T) {
		m.storage.EXPECT().BatchUpdate(gomock.Any(), []models.Metrics{
			{MType: "gauge", ID: "gauge1", Value: utils.FloatPtr(1.23)},
			{MType: "counter", ID: "counter1", Delta: utils.IntPtr(10)},
		}).Return(fmt.Errorf("failed to batch update metrics"))

		req := &proto.UpdateMetricsRequest{
			Metrics: []*proto.Metric{
				{Id: "gauge1", Type: "gauge", Value: 1.23},
				{Id: "counter1", Type: "counter", Delta: 10},
			},
		}
		_, err := srv.UpdateMetrics(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.Internal, st.Code(), "Expected error code Internal")
	})
	t.Run("GetNonExistentMetricType", func(t *testing.T) {
		req := &proto.UpdateMetricsRequest{
			Metrics: []*proto.Metric{
				{Id: "gauge1", Type: "gauge1"},
				{Id: "counter1", Type: "counter1"},
			},
		}
		_, err := srv.UpdateMetrics(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.NotFound, st.Code(), "Expected error code NotFound")
	})
}

func TestGetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := &mocks{
		storage: storagemock.NewMockStorage(ctrl),
	}

	srv := &MetricsServer{
		storage: m.storage,
	}

	reqUpdate := &proto.UpdateMetricsRequest{
		Metrics: []*proto.Metric{
			{Id: "gauge1", Type: "gauge", Value: 1.23},
			{Id: "counter1", Type: "counter", Delta: 10},
		},
	}
	m.storage.EXPECT().BatchUpdate(gomock.Any(), []models.Metrics{
		{MType: "gauge", ID: "gauge1", Value: utils.FloatPtr(1.23)},
		{MType: "counter", ID: "counter1", Delta: utils.IntPtr(10)},
	})

	_, err := srv.UpdateMetrics(context.Background(), reqUpdate)
	assert.NoError(t, err)

	t.Run("GetAllMetrics_Success", func(t *testing.T) {
		m.storage.EXPECT().
			GetAllGauges(gomock.Any()).
			Return([]storage.GaugeMetric{
				{Name: "gauge1", Value: 1.23},
			}, nil).
			Times(1)

		m.storage.EXPECT().
			GetAllCounters(gomock.Any()).
			Return([]storage.CounterMetric{
				{Name: "counter1", Value: 10},
			}, nil).
			Times(1)

		req := &emptypb.Empty{}
		resp, err := srv.GetAllMetrics(context.Background(), req)
		assert.NoError(t, err)
		assert.Len(t, resp.Metrics, 2)
		assert.Equal(t, "gauge1", resp.Metrics[0].Id)
		assert.Equal(t, "counter1", resp.Metrics[1].Id)
	})

	t.Run("GetAllMetrics_GetAllGaugesError", func(t *testing.T) {
		m.storage.EXPECT().
			GetAllGauges(gomock.Any()).
			Return(nil, fmt.Errorf("storage error")).
			Times(1)

		req := &emptypb.Empty{}
		_, err := srv.GetAllMetrics(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "Error fetching gauge metrics: storage error", st.Message())
	})

	t.Run("GetAllMetrics_GetAllCountersError", func(t *testing.T) {
		m.storage.EXPECT().
			GetAllGauges(gomock.Any()).
			Return([]storage.GaugeMetric{
				{Name: "gauge1", Value: 1.23},
			}, nil).
			Times(1)

		m.storage.EXPECT().
			GetAllCounters(gomock.Any()).
			Return(nil, fmt.Errorf("storage error")).
			Times(1)

		req := &emptypb.Empty{}
		_, err := srv.GetAllMetrics(context.Background(), req)
		assert.Error(t, err)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		assert.Equal(t, "Error fetching counter metrics: storage error", st.Message())
	})
}

func TestNewMetricsServer(t *testing.T) {
	s, _ := memory.NewMemStorage(context.Background(), 300, "/tmp/metrics-db.json", false)
	server := NewMetricsServer(s)

	assert.NotNil(t, server, "Сервер не должен быть nil")
}

func TestStartGRPCServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := &mocks{
		storage: storagemock.NewMockStorage(ctrl),
	}

	s := &MetricsServer{
		Address:       "127.0.0.1:50051",
		TrustedSubnet: "127.0.0.1/32",
		Logger:        nil,
		PrivateKey:    nil,
		HashKey:       "testkey",
	}

	go func() {
		s.StartGRPCServer(m.storage)
	}()

	require.NotNil(t, s)
}
