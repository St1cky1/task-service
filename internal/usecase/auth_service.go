package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/infrastructure/auth"
	"github.com/St1cky1/task-service/internal/repository"
)

type AuthService struct {
	userRepo         repository.IUserRepository
	refreshTokenRepo repository.IRefreshTokenRepository
	passwordManager  *auth.PasswordManager
	jwtManager       *auth.JWTManager
}

func NewAuthService(
	userRepo repository.IUserRepository,
	refreshTokenRepo repository.IRefreshTokenRepository,
	passwordManager *auth.PasswordManager,
	jwtManager *auth.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordManager:  passwordManager,
		jwtManager:       jwtManager,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req *entity.RegisterRequest) (*entity.LoginResponse, error) {
	// Проверяем, что пользователь с таким email не существует
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Хешируем пароль
	passwordHash, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	user, err := s.userRepo.CreateWithAuth(ctx, req.Name, req.Email, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токены
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Сохраняем хеш refresh token в БД
	refreshTokenHash := s.hashToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.refreshTokenRepo.Save(ctx, user.ID, refreshTokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Обновляем last_login
	updates := make(map[string]interface{})
	updates["last_login"] = time.Now()
	_, err = s.userRepo.Update(ctx, user.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update last_login: %w", err)
	}

	return &entity.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login логинит пользователя
func (s *AuthService) Login(ctx context.Context, req *entity.LoginRequest) (*entity.LoginResponse, error) {
	// Ищем пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Проверяем активность пользователя
	if !user.IsActive {
		return nil, fmt.Errorf("user is not active")
	}

	// Проверяем пароль
	if !s.passwordManager.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Генерируем токены
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Сохраняем хеш refresh token в БД
	refreshTokenHash := s.hashToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.refreshTokenRepo.Save(ctx, user.ID, refreshTokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Обновляем last_login
	updates := make(map[string]interface{})
	updates["last_login"] = time.Now()
	_, err = s.userRepo.Update(ctx, user.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update last_login: %w", err)
	}

	return &entity.LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken обновляет access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (*entity.RefreshTokenResponse, error) {
	// Проверяем refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Проверяем, есть ли этот токен в БД
	refreshTokenHash := s.hashToken(refreshTokenStr)
	storedToken, err := s.refreshTokenRepo.GetByHash(ctx, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if storedToken == nil {
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// Генерируем новый access token
	newAccessToken, err := s.jwtManager.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Генерируем новый refresh token
	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(claims.UserID, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Откатываем старый refresh token
	err = s.refreshTokenRepo.Revoke(ctx, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Сохраняем новый refresh token
	newRefreshTokenHash := s.hashToken(newRefreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	err = s.refreshTokenRepo.Save(ctx, claims.UserID, newRefreshTokenHash, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return &entity.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout откатывает все refresh токены пользователя
func (s *AuthService) Logout(ctx context.Context, userID int) error {
	err := s.refreshTokenRepo.RevokeAll(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens: %w", err)
	}
	return nil
}

// hashToken генерирует хеш токена для хранения в БД
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
