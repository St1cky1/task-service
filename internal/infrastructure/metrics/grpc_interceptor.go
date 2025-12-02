package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor возвращает unary interceptor для gRPC метрик
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Вызываем handler
	resp, err := handler(ctx, req)

	// Записываем метрики
	duration := time.Since(start).Seconds()
	metrics := GetMetrics()

	metrics.GRPCRequestsTotal.Inc()
	metrics.GRPCRequestDuration.Observe(duration)

	// Считаем ошибки
	if err != nil {
		metrics.GRPCErrors.Inc()
	}

	return resp, err
}

// StreamServerInterceptor возвращает stream interceptor для gRPC метрик
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()

	// Вызываем handler
	err := handler(srv, ss)

	// Записываем метрики
	duration := time.Since(start).Seconds()
	metrics := GetMetrics()

	metrics.GRPCRequestsTotal.Inc()
	metrics.GRPCRequestDuration.Observe(duration)

	// Считаем ошибки
	if err != nil && status.Code(err) != codes.OK {
		metrics.GRPCErrors.Inc()
	}

	return err
}