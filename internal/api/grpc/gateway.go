package grpc

import (
	"context"
	"net"
	"net/http"

	"github.com/St1cky1/task-service/internal/usecase"
	pb "github.com/St1cky1/task-service/proto/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server представляет gRPC сервер с поддержкой Gateway
type Server struct {
	grpcServer  *grpc.Server
	taskService *usecase.TaskService
	userService *usecase.UserService
	authService *usecase.AuthService
}

// NewGRPCServer создает новый gRPC сервер
func NewGRPCServer(taskService *usecase.TaskService, userService *usecase.UserService, authService *usecase.AuthService) *Server {
	return &Server{
		grpcServer:  grpc.NewServer(),
		taskService: taskService,
		userService: userService,
		authService: authService,
	}
}

// Start запускает gRPC сервер на указанном порту
func (s *Server) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// Регистрируем TaskService
	taskHandler := NewTaskServiceServer(s.taskService)
	pb.RegisterTaskServiceServer(s.grpcServer, taskHandler)

	// Регистрируем UserService
	userHandler := NewUserServiceServer(s.userService, s.authService)
	pb.RegisterUserServiceServer(s.grpcServer, userHandler)

	return s.grpcServer.Serve(listener)
}

// Stop останавливает gRPC сервер
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// StartGateway запускает gRPC Gateway на указанном порту
func (s *Server) StartGateway(ctx context.Context, grpcPort, gatewayPort string) error {
	mux := runtime.NewServeMux()

	// Подключаемся к gRPC серверу
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterTaskServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		return err
	}

	err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts)
	if err != nil {
		return err
	}

	// Запускаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + gatewayPort,
		Handler: mux,
	}

	return server.ListenAndServe()
}
