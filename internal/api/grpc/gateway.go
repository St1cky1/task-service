package grpc

import (
	"context"
	"net"
	"net/http"

	"github.com/St1cky1/task-service/internal/infrastructure/metrics"
	"github.com/St1cky1/task-service/internal/usecase"
	pb "github.com/St1cky1/task-service/proto/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		grpcServer: grpc.NewServer(
			grpc.UnaryInterceptor(metrics.UnaryServerInterceptor),
			grpc.StreamInterceptor(metrics.StreamServerInterceptor),
		),
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

	// Создаем новый router с middleware и metrics endpoint
	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.Handle("/", metrics.HTTPMiddleware(corsMiddleware(mux)))

	// Запускаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + gatewayPort,
		Handler: httpMux,
	}

	return server.ListenAndServe()
}

// corsMiddleware добавляет CORS headers для REST клиентов (включая Flutter)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
