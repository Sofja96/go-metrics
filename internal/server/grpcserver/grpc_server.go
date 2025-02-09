package grpcserver

import (
	"context"
	"crypto/rsa"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Sofja96/go-metrics.git/internal/models"
	"github.com/Sofja96/go-metrics.git/internal/proto"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
)

type MetricsServer struct {
	proto.UnimplementedMetricsServer
	storage       storage.Storage
	server        *grpc.Server
	Address       string
	TrustedSubnet string
	Logger        *zap.SugaredLogger
	PrivateKey    *rsa.PrivateKey
	HashKey       string
}

func NewMetricsServer(storage storage.Storage) *MetricsServer {
	return &MetricsServer{storage: storage}
}

func (s *MetricsServer) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	mName := req.GetName()
	mType := req.GetType()

	resp := &proto.GetMetricResponse{
		Metric: &proto.Metric{},
	}

	switch mType {
	case "gauge":
		value, ok := s.storage.GetGaugeValue(ctx, mName)
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Gauge metric '%s' not found", mName)
		}
		resp.Metric.Value = value
	case "counter":
		value, ok := s.storage.GetCounterValue(ctx, mName)
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Counter metric '%s' not found", mName)
		}
		resp.Metric.Delta = value
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Invalid metric type '%s'. Metric type can only be 'gauge' or 'counter'", mType)
	}

	return resp, nil
}

func (s *MetricsServer) GetAllMetrics(ctx context.Context, _ *emptypb.Empty) (*proto.GetAllMetricsResponse, error) {
	resp := &proto.GetAllMetricsResponse{
		Metrics: []*proto.Metric{},
	}

	gauges, err := s.storage.GetAllGauges(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error fetching gauge metrics: %v", err)
	}

	counters, err := s.storage.GetAllCounters(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error fetching counter metrics: %v", err)
	}

	for _, gauge := range gauges {
		resp.Metrics = append(resp.Metrics, &proto.Metric{
			Id:    gauge.Name,
			Type:  "gauge",
			Value: gauge.Value,
		})
	}

	for _, counter := range counters {
		resp.Metrics = append(resp.Metrics, &proto.Metric{
			Id:    counter.Name,
			Type:  "counter",
			Delta: counter.Value,
		})
	}

	return resp, nil
}

func (s *MetricsServer) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	metric := req.GetMetric()

	switch metric.GetType() {
	case "gauge":
		_, err := s.storage.UpdateGauge(ctx, metric.GetId(), metric.GetValue())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Error updating gauge metric '%s': %v", metric.Id, err)
		}
	case "counter":
		_, err := s.storage.UpdateCounter(ctx, metric.GetId(), metric.GetDelta())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Error updating counter metric '%s': %v", metric.Id, err)
		}
	default:
		return nil, status.Errorf(codes.NotFound, "unsupported metric type: %s", metric.GetType())
	}
	return &proto.UpdateMetricResponse{Success: true}, nil
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	var metrics []models.Metrics
	for _, protoMetric := range req.GetMetrics() {
		metric := models.Metrics{
			ID:    protoMetric.GetId(),
			MType: protoMetric.GetType(),
		}

		switch protoMetric.GetType() {
		case "counter":
			delta := protoMetric.GetDelta()
			metric.Delta = &delta
		case "gauge":
			value := protoMetric.GetValue()
			metric.Value = &value
		default:
			return nil, status.Errorf(codes.NotFound, "unsupported metric type: %s", protoMetric.GetType())
		}

		metrics = append(metrics, metric)
	}

	err := s.storage.BatchUpdate(ctx, metrics)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to batch update metrics: %v", err)
	}

	return &proto.UpdateMetricsResponse{Success: true}, nil
}

func (s *MetricsServer) StartGRPCServer(store storage.Storage) {
	lis, err := net.Listen("tcp", s.Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor(s.Logger),
			ValidateTrustedSubnetInterceptor(s.TrustedSubnet, s.Logger),
			HMACInterceptor(s.Logger, []byte(s.HashKey)),
			DecryptInterceptor(s.Logger, s.PrivateKey),
			GzipInterceptor(s.Logger),
		),
	)
	proto.RegisterMetricsServer(grpcServer, NewMetricsServer(store))

	reflection.Register(grpcServer)

	log.Printf("gRPC server listening at %v", s.Address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
