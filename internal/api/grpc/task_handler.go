package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/infrastructure/auth"
	"github.com/St1cky1/task-service/internal/usecase"
	pb "github.com/St1cky1/task-service/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TaskServiceServer реализует gRPC TaskService
type TaskServiceServer struct {
	pb.UnimplementedTaskServiceServer
	taskService *usecase.TaskService
	jwtManager  *auth.JWTManager
}

// NewTaskServiceServer создает новый TaskServiceServer
func NewTaskServiceServer(taskService *usecase.TaskService, jwtManager *auth.JWTManager) *TaskServiceServer {
	return &TaskServiceServer{
		taskService: taskService,
		jwtManager:  jwtManager,
	}
}

// extractUserIDFromContext извлекает userID из JWT токена в метаданных
func (s *TaskServiceServer) extractUserIDFromContext(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("no metadata found")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return 0, fmt.Errorf("no authorization header")
	}

	authHeader := authHeaders[0]
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return 0, fmt.Errorf("invalid authorization header format")
	}

	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, nil
}

// CreateTask создает новую задачу
func (s *TaskServiceServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	taskReq := &entity.CreateTaskRequest{
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.TaskStatus(req.Status),
		OwnerId:     int(req.OwnerId),
	}

	task, err := s.taskService.CreateTask(ctx, taskReq, int(req.OwnerId))
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		case entity.ErrInvalidTaskData:
			return nil, status.Error(codes.InvalidArgument, "invalid task data")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		OwnerId:     int32(task.OwnerId),
		CreatedAt:   task.CreatedAt.String(),
		UpdatedAt:   task.UpdatedAt.String(),
	}, nil
}

// GetTask получает задачу по ID
func (s *TaskServiceServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
	userID, err := s.extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	task, err := s.taskService.GetTask(ctx, int(req.Id), userID)
	if err != nil {
		switch err {
		case entity.ErrTaskNotFound:
			return nil, status.Error(codes.NotFound, "task not found")
		case entity.ErrForbidden:
			return nil, status.Error(codes.PermissionDenied, "access denied")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		OwnerId:     int32(task.OwnerId),
		CreatedAt:   task.CreatedAt.String(),
		UpdatedAt:   task.UpdatedAt.String(),
	}, nil
}

// UpdateTask обновляет задачу
func (s *TaskServiceServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.TaskResponse, error) {
	userID, err := s.extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	updateReq := &entity.UpdateTaskRequest{
		Title:       req.Title,
		Status:      entity.TaskStatus(req.Status),
		Description: req.Description,
	}

	task, err := s.taskService.UpdateTask(ctx, int(req.Id), userID, updateReq)
	if err != nil {
		switch err {
		case entity.ErrTaskNotFound:
			return nil, status.Error(codes.NotFound, "task not found")
		case entity.ErrNoFieldsToUpdate:
			return nil, status.Error(codes.InvalidArgument, "no fields to update")
		case entity.ErrForbidden:
			return nil, status.Error(codes.PermissionDenied, "access denied")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		OwnerId:     int32(task.OwnerId),
		CreatedAt:   task.CreatedAt.String(),
		UpdatedAt:   task.UpdatedAt.String(),
	}, nil
}

// DeleteTask удаляет задачу
func (s *TaskServiceServer) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	userID, err := s.extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	err = s.taskService.DeleteTask(ctx, int(req.Id), userID)
	if err != nil {
		switch err {
		case entity.ErrTaskNotFound:
			return nil, status.Error(codes.NotFound, "task not found")
		case entity.ErrForbidden:
			return nil, status.Error(codes.PermissionDenied, "access denied")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.DeleteTaskResponse{Success: true}, nil
}

// ListTasks получает список задач
func (s *TaskServiceServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	userID, err := s.extractUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	tasks, err := s.taskService.ListTasks(ctx, userID, req.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTasks := make([]*pb.TaskResponse, len(tasks))
	for i, task := range tasks {
		pbTasks[i] = &pb.TaskResponse{
			Id:          int32(task.ID),
			Title:       task.Title,
			Description: task.Description,
			Status:      string(task.Status),
			OwnerId:     int32(task.OwnerId),
			CreatedAt:   task.CreatedAt.String(),
			UpdatedAt:   task.UpdatedAt.String(),
		}
	}

	return &pb.ListTasksResponse{Tasks: pbTasks}, nil
}
