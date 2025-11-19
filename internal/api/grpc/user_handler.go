package grpc

import (
	"context"
	"io"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/usecase"
	pb "github.com/St1cky1/task-service/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServiceServer реализует gRPC UserService
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userService *usecase.UserService
	authService *usecase.AuthService
}

// NewUserServiceServer создает новый UserServiceServer
func NewUserServiceServer(userService *usecase.UserService, authService *usecase.AuthService) *UserServiceServer {
	return &UserServiceServer{
		userService: userService,
		authService: authService,
	}
}

// CreateUser создает нового пользователя
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	userReq := &entity.CreateUserRequest{
		Name: req.Name,
	}

	user, err := s.userService.CreateUser(ctx, userReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.UserResponse{
		Id:        int32(user.ID),
		Name:      user.Name,
		AvatarUrl: avatarURL,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}, nil
}

// GetUser получает пользователя по ID
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	user, err := s.userService.GetUser(ctx, int(req.Id))
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.UserResponse{
		Id:        int32(user.ID),
		Name:      user.Name,
		AvatarUrl: avatarURL,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}, nil
}

// UpdateUser обновляет пользователя
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	userReq := &entity.UpdateUserRequest{
		Name: req.Name,
	}

	user, err := s.userService.UpdateUser(ctx, int(req.Id), userReq)
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.UserResponse{
		Id:        int32(user.ID),
		Name:      user.Name,
		AvatarUrl: avatarURL,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
	}, nil
}

// DeleteUser удаляет пользователя
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, int(req.Id))
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

// ListUsers получает список пользователей
func (s *UserServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	users, err := s.userService.ListUsers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbUsers := make([]*pb.UserResponse, len(users))
	for i, user := range users {
		avatarURL := ""
		if user.AvatarURL != nil {
			avatarURL = *user.AvatarURL
		}

		pbUsers[i] = &pb.UserResponse{
			Id:        int32(user.ID),
			Name:      user.Name,
			AvatarUrl: avatarURL,
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
		}
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Total: int32(len(pbUsers)),
	}, nil
}

// UploadAvatar загружает аватарку пользователя (клиентский stream)
func (s *UserServiceServer) UploadAvatar(stream pb.UserService_UploadAvatarServer) error {
	// Читаем первый пакет для получения user_id
	firstMsg, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return status.Error(codes.InvalidArgument, "empty stream")
		}
		return status.Error(codes.Internal, err.Error())
	}

	userID := int(firstMsg.UserId)
	contentType := firstMsg.ContentType
	var data []byte
	data = append(data, firstMsg.Data...)

	// Читаем остальные пакеты
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Error(codes.Internal, err.Error())
		}
		data = append(data, msg.Data...)
	}

	// Загружаем аватарку
	filePath, err := s.userService.UploadAvatar(stream.Context(), userID, data, contentType)
	if err != nil {
		switch err {
		case entity.ErrUserNotFound:
			return status.Error(codes.NotFound, "user not found")
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	// Отправляем ответ
	return stream.SendAndClose(&pb.UploadAvatarResponse{
		Success:   true,
		Message:   "Avatar uploaded successfully",
		AvatarUrl: filePath,
	})
}

// DownloadAvatar скачивает аватарку пользователя (серверный stream)
func (s *UserServiceServer) DownloadAvatar(req *pb.DownloadAvatarRequest, stream pb.UserService_DownloadAvatarServer) error {
	userID := int(req.UserId)

	// Используем stream метод из UserService
	dataChan, errChan := s.userService.DownloadAvatarStream(stream.Context(), userID, 64*1024) // 64KB chunks

	// Отправляем чанки
	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				return nil
			}
			if err := stream.Send(&pb.DownloadAvatarResponse{
				Data: data,
			}); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		case err := <-errChan:
			if err != nil {
				switch err {
				case entity.ErrUserNotFound:
					return status.Error(codes.NotFound, "user not found")
				default:
					return status.Error(codes.Internal, err.Error())
				}
			}
			return nil
		}
	}
}
