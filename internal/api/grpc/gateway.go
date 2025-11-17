package grpc

import (
	"context"
	"fmt"
	"net/http"

	pb "github.com/St1cky1/task-service/proto/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGatewayHandler создает HTTP Gateway для gRPC
func NewGatewayHandler(ctx context.Context, grpcAddr string) (http.Handler, error) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := pb.RegisterTaskServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	return mux, nil
}
