package grpc

import (
	"context"
	"fmt"

	"github.com/St1cky1/task-service/internal/entity"
	pb "github.com/St1cky1/task-service/proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register регистрирует нового пользователя
func (s *UserServiceServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email and password are required")
	}

	registerReq := &entity.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	loginResp, err := s.authService.Register(ctx, registerReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertRegisterResponse(loginResp), nil
}

// Login логинит пользователя
func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	loginReq := &entity.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	loginResp, err := s.authService.Login(ctx, loginReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertLoginResponse(loginResp), nil
}

// RefreshToken обновляет access token
func (s *UserServiceServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	refreshResp, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  refreshResp.AccessToken,
		RefreshToken: refreshResp.RefreshToken,
	}, nil
}

// Logout откатывает все refresh токены пользователя
func (s *UserServiceServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.authService.Logout(ctx, int(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to logout: %v", err))
	}

	return &pb.LogoutResponse{
		Success: true,
		Message: "logged out successfully",
	}, nil
}

// convertLoginResponse конвертирует entity.LoginResponse в pb.LoginResponse
func convertLoginResponse(resp *entity.LoginResponse) *pb.LoginResponse {
	return &pb.LoginResponse{
		User:         convertUser(resp.User),
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}

// convertRegisterResponse конвертирует entity.LoginResponse в pb.RegisterResponse
func convertRegisterResponse(resp *entity.LoginResponse) *pb.RegisterResponse {
	return &pb.RegisterResponse{
		User:         convertUser(resp.User),
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}

// convertUser конвертирует entity.User в pb.UserResponse
func convertUser(user *entity.User) *pb.UserResponse {
	var lastLogin string
	if user.LastLogin != nil {
		lastLogin = user.LastLogin.Format("2006-01-02T15:04:05Z07:00")
	}

	var email string
	if user.Email != nil {
		email = *user.Email
	}

	var avatarURL string
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	return &pb.UserResponse{
		Id:        int32(user.ID),
		Name:      user.Name,
		Email:     email,
		AvatarUrl: avatarURL,
		IsActive:  user.IsActive,
		LastLogin: lastLogin,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
