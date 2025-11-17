package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/usecase"
	pb "github.com/St1cky1/task-service/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedUserServiceServer
	taskService *usecase.TaskService
	userService *usecase.UserService
	server      *grpc.Server
}

func NewGRPCServer(taskService *usecase.TaskService, userService *usecase.UserService) *GRPCServer {
	return &GRPCServer{
		taskService: taskService,
		userService: userService,
	}
}

func (s *GRPCServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
	)
	// Регистрируем оба сервиса на одном gRPC сервере
	pb.RegisterTaskServiceServer(s.server, s)
	pb.RegisterUserServiceServer(s.server, s)
	reflection.Register(s.server)

	log.Printf("gRPC server listening on :%s", port)
	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

func (s *GRPCServer) unaryInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Printf("gRPC method: %s", info.FullMethod)
	return handler(ctx, req)
}

// CreateTask - создание задачи
func (s *GRPCServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	createReq := &entity.CreateTaskRequest{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Status:      entity.TaskStatus(req.GetStatus()),
		OwnerId:     int(req.GetOwnerId()),
	}

	task, err := s.taskService.CreateTask(ctx, createReq, int(req.GetOwnerId()))
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

	return s.taskToProto(task), nil
}

// GetTask - получение задачи
func (s *GRPCServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
	task, err := s.taskService.GetTask(ctx, int(req.GetId()), 1) // TODO: получить userID из контекста
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

	return s.taskToProto(task), nil
}

// UpdateTask - обновление задачи
func (s *GRPCServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.TaskResponse, error) {
	updateReq := entity.UpdateTaskRequest{
		Title:  req.GetTitle(),
		Status: entity.TaskStatus(req.GetStatus()),
	}

	if req.Description != nil {
		desc := req.GetDescription()
		updateReq.Description = &desc
	}

	task, err := s.taskService.UpdateTask(ctx, int(req.GetId()), 1, &updateReq) // TODO: получить userID из контекста
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

	return s.taskToProto(task), nil
}

// DeleteTask - удаление задачи
func (s *GRPCServer) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	err := s.taskService.DeleteTask(ctx, int(req.GetId()), 1) // TODO: получить userID из контекста
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

// ListTasks - список задач
func (s *GRPCServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	tasks, err := s.taskService.ListTasks(ctx, 1, req.GetStatus()) // TODO: получить userID из контекста
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoTasks := make([]*pb.TaskResponse, len(tasks))
	for i, task := range tasks {
		protoTasks[i] = s.taskToProto(&task)
	}

	return &pb.ListTasksResponse{Tasks: protoTasks}, nil
}

// Вспомогательный метод для преобразования entity.Task в pb.TaskResponse
func (s *GRPCServer) taskToProto(task *entity.Task) *pb.TaskResponse {
	return &pb.TaskResponse{
		Id:          int32(task.ID),
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		OwnerId:     int32(task.OwnerId),
		CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt.Format(time.RFC3339),
	}
}

// ===== UserService Methods =====

// CreateUser - создание пользователя
func (s *GRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	createReq := &entity.CreateUserRequest{
		Name: req.GetName(),
	}

	user, err := s.userService.CreateUser(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.userToProto(user), nil
}

// GetUser - получение пользователя
func (s *GRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.GetUser(ctx, int(req.GetId()))
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return s.userToProto(user), nil
}

// UpdateUser - обновление пользователя
func (s *GRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	updateReq := entity.UpdateUserRequest{
		Name: req.GetName(),
	}

	user, err := s.userService.UpdateUser(ctx, int(req.GetId()), &updateReq)
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return s.userToProto(user), nil
}

// DeleteUser - удаление пользователя
func (s *GRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, int(req.GetId()))
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.DeleteUserResponse{Success: true}, nil
}

// ListUsers - список пользователей
func (s *GRPCServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.userService.ListUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoUsers := make([]*pb.UserResponse, len(users))
	for i, user := range users {
		protoUsers[i] = s.userToProto(&user)
	}

	return &pb.ListUsersResponse{
		Users: protoUsers,
		Total: int32(len(users)),
	}, nil
}

// UploadAvatar - загрузка аватарки (streaming)
func (s *GRPCServer) UploadAvatar(stream pb.UserService_UploadAvatarServer) error {
	ctx := stream.Context()

	var userID int32
	var data []byte
	var contentType string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("❌ Error receiving upload avatar stream: %v", err)
			return status.Error(codes.Internal, "error receiving data")
		}

		if req.GetUserId() != 0 {
			userID = req.GetUserId()
		}
		if req.GetContentType() != "" {
			contentType = req.GetContentType()
		}

		data = append(data, req.GetData()...)
	}

	if userID == 0 {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	if len(data) == 0 {
		return status.Error(codes.InvalidArgument, "avatar data is empty")
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Загружаем аватарку
	filePath, err := s.userService.UploadAvatar(ctx, int(userID), data, contentType)
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return status.Error(codes.NotFound, "user not found")
		default:
			log.Printf("❌ Error uploading avatar: %v", err)
			return status.Error(codes.Internal, err.Error())
		}
	}

	resp := &pb.UploadAvatarResponse{
		Success:   true,
		Message:   "avatar uploaded successfully",
		AvatarUrl: filePath,
	}

	if err := stream.SendAndClose(resp); err != nil {
		log.Printf("❌ Error sending upload response: %v", err)
		return status.Error(codes.Internal, "error sending response")
	}

	return nil
}

// DownloadAvatar - скачивание аватарки (streaming)
func (s *GRPCServer) DownloadAvatar(req *pb.DownloadAvatarRequest, stream pb.UserService_DownloadAvatarServer) error {
	ctx := stream.Context()
	userID := int(req.GetUserId())

	dataChan, errChan := s.userService.DownloadAvatarStream(ctx, userID, 64*1024) // 64KB chunks

	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				return nil
			}
			resp := &pb.DownloadAvatarResponse{
				Data: data,
			}
			if err := stream.Send(resp); err != nil {
				log.Printf("❌ Error sending avatar data: %v", err)
				return status.Error(codes.Internal, "error sending data")
			}

		case err := <-errChan:
			if err != nil {
				switch err {
				case entity.ErrUserNotFound:
					return status.Error(codes.NotFound, "user not found")
				default:
					if err.Error() == "avatar not found" {
						return status.Error(codes.NotFound, "avatar not found")
					}
					log.Printf("❌ Error downloading avatar: %v", err)
					return status.Error(codes.Internal, err.Error())
				}
			}

		case <-ctx.Done():
			return status.Error(codes.Canceled, "request canceled")
		}
	}
}

// Вспомогательный метод для преобразования entity.User в pb.UserResponse
func (s *GRPCServer) userToProto(user *entity.User) *pb.UserResponse {
	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.UserResponse{
		Id:        int32(user.ID),
		Name:      user.Name,
		AvatarUrl: avatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
