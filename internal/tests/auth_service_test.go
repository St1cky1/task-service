package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/infrastructure/auth"
	"github.com/St1cky1/task-service/internal/repository"
	"github.com/St1cky1/task-service/internal/usecase"
)

// MockUserRepository для тестирования AuthService
type MockUserRepositoryAuth struct {
	GetByEmailFunc     func(ctx context.Context, email string) (*entity.User, error)
	CreateWithAuthFunc func(ctx context.Context, name, email, passwordHash string) (*entity.User, error)
	UpdateFunc         func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error)
	GetByIdFunc        func(ctx context.Context, id int) (*entity.User, error)
	CreateFunc         func(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error)
	ListFunc           func(ctx context.Context) ([]entity.User, error)
	DeleteFunc         func(ctx context.Context, id int) error
}

var _ repository.IUserRepository = (*MockUserRepositoryAuth)(nil)

func (m *MockUserRepositoryAuth) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepositoryAuth) CreateWithAuth(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
	if m.CreateWithAuthFunc != nil {
		return m.CreateWithAuthFunc(ctx, name, email, passwordHash)
	}
	return nil, nil
}

func (m *MockUserRepositoryAuth) Create(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil, nil
}

func (m *MockUserRepositoryAuth) List(ctx context.Context) ([]entity.User, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserRepositoryAuth) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepositoryAuth) Update(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates)
	}
	return nil, nil
}

func (m *MockUserRepositoryAuth) GetById(ctx context.Context, id int) (*entity.User, error) {
	if m.GetByIdFunc != nil {
		return m.GetByIdFunc(ctx, id)
	}
	return nil, nil
}

// MockRefreshTokenRepository для тестирования AuthService
type MockRefreshTokenRepository struct {
	SaveFunc           func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	GetByHashFunc      func(ctx context.Context, tokenHash string) (*repository.RefreshToken, error)
	RevokeFunc         func(ctx context.Context, tokenHash string) error
	RevokeAllFunc      func(ctx context.Context, userID int) error
	GetByUserIDFunc    func(ctx context.Context, userID int) ([]repository.RefreshToken, error)
	CleanupExpiredFunc func(ctx context.Context) error
}

var _ repository.IRefreshTokenRepository = (*MockRefreshTokenRepository)(nil)

func (m *MockRefreshTokenRepository) Save(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, userID, tokenHash, expiresAt)
	}
	return nil
}

func (m *MockRefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*repository.RefreshToken, error) {
	if m.GetByHashFunc != nil {
		return m.GetByHashFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(ctx, tokenHash)
	}
	return nil
}

func (m *MockRefreshTokenRepository) RevokeAll(ctx context.Context, userID int) error {
	if m.RevokeAllFunc != nil {
		return m.RevokeAllFunc(ctx, userID)
	}
	return nil
}

func (m *MockRefreshTokenRepository) GetByUserID(ctx context.Context, userID int) ([]repository.RefreshToken, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRefreshTokenRepository) CleanupExpired(ctx context.Context) error {
	if m.CleanupExpiredFunc != nil {
		return m.CleanupExpiredFunc(ctx)
	}
	return nil
}

// Tests for Register

func TestRegisterSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	expectedPassword := "password123"
	passwordHash, _ := passwordManager.HashPassword(expectedPassword)

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
		Email: func() *string {
			e := "test@example.com"
			return &e
		}(),
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return nil, nil // User doesn't exist
		},
		CreateWithAuthFunc: func(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
			return mockUser, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		SaveFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: expectedPassword,
	}

	result, err := service.Register(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.User.ID != mockUser.ID {
		t.Errorf("Expected user ID %d, got %d", mockUser.ID, result.User.ID)
	}

	if result.User.Name != mockUser.Name {
		t.Errorf("Expected user name %s, got %s", mockUser.Name, result.User.Name)
	}

	if result.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if result.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
}

func TestRegisterUserAlreadyExists(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	existingUser := &entity.User{
		ID:       1,
		Name:     "Existing User",
		Email:    func() *string { e := "test@example.com"; return &e }(),
		IsActive: true,
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return existingUser, nil // User already exists
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	result, err := service.Register(ctx, req)
	if err == nil {
		t.Fatal("Expected error when user already exists")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err.Error() != "user with this email already exists" {
		t.Errorf("Expected 'user with this email already exists' error, got %v", err)
	}
}

func TestRegisterPasswordHashingError(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
		Email: func() *string {
			e := "test@example.com"
			return &e
		}(),
		IsActive: true,
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return nil, nil
		},
		CreateWithAuthFunc: func(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
			return mockUser, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		SaveFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	result, _ := service.Register(ctx, req)
	if result == nil {
		t.Errorf("Expected non-nil result on success")
	}
}

func TestRegisterRefreshTokenSaveError(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
		Email: func() *string {
			e := "test@example.com"
			return &e
		}(),
		IsActive: true,
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return nil, nil
		},
		CreateWithAuthFunc: func(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		SaveFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return fmt.Errorf("database error")
		},
	}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	result, err := service.Register(ctx, req)
	if err == nil {
		t.Fatal("Expected error when refresh token save fails")
	}

	if result != nil {
		t.Errorf("Expected nil result")
	}
}

// Tests for Login

func TestLoginSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	expectedPassword := "password123"
	passwordHash, _ := passwordManager.HashPassword(expectedPassword)

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
		Email: func() *string {
			e := "test@example.com"
			return &e
		}(),
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return mockUser, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		SaveFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.LoginRequest{
		Email:    "test@example.com",
		Password: expectedPassword,
	}

	result, err := service.Login(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.User.ID != mockUser.ID {
		t.Errorf("Expected user ID %d, got %d", mockUser.ID, result.User.ID)
	}

	if result.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if result.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
}

func TestLoginUserNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return nil, nil // User not found
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	result, err := service.Login(ctx, req)
	if err == nil {
		t.Fatal("Expected error when user not found")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err.Error() != "invalid email or password" {
		t.Errorf("Expected 'invalid email or password' error, got %v", err)
	}
}

func TestLoginUserNotActive(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	inactiveUser := &entity.User{
		ID:   1,
		Name: "Test User",
		Email: func() *string {
			e := "test@example.com"
			return &e
		}(),
		PasswordHash: "somehash",
		IsActive:     false, // User is not active
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return inactiveUser, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	result, err := service.Login(ctx, req)
	if err == nil {
		t.Fatal("Expected error when user is not active")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err.Error() != "user is not active" {
		t.Errorf("Expected 'user is not active' error, got %v", err)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	correctPassword := "correct_password"
	passwordHash, _ := passwordManager.HashPassword(correctPassword)

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
		Email: func() *string {
			e := "test@example.com"
			return &e
		}(),
		PasswordHash: passwordHash,
		IsActive:     true,
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByEmailFunc: func(ctx context.Context, email string) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	req := &entity.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong_password",
	}

	result, err := service.Login(ctx, req)
	if err == nil {
		t.Fatal("Expected error with invalid password")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err.Error() != "invalid email or password" {
		t.Errorf("Expected 'invalid email or password' error, got %v", err)
	}
}

// Tests for RefreshToken

func TestRefreshTokenSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	// Generate a valid refresh token
	refreshToken, _ := jwtManager.GenerateRefreshToken(1, "test@example.com")

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		GetByHashFunc: func(ctx context.Context, tokenHash string) (*repository.RefreshToken, error) {
			return &repository.RefreshToken{
				UserID:    1,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
		RevokeFunc: func(ctx context.Context, tokenHash string) error {
			return nil
		},
		SaveFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}

	mockUserRepo := &MockUserRepositoryAuth{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	result, err := service.RefreshToken(ctx, refreshToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if result.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
}

func TestRefreshTokenInvalidToken(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}
	mockUserRepo := &MockUserRepositoryAuth{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	result, err := service.RefreshToken(ctx, "invalid_token")
	if err == nil {
		t.Fatal("Expected error with invalid token")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

func TestRefreshTokenNotFoundInDB(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	// Generate a valid refresh token
	refreshToken, _ := jwtManager.GenerateRefreshToken(1, "test@example.com")

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		GetByHashFunc: func(ctx context.Context, tokenHash string) (*repository.RefreshToken, error) {
			return nil, nil // Token not found in DB
		},
	}

	mockUserRepo := &MockUserRepositoryAuth{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	result, err := service.RefreshToken(ctx, refreshToken)
	if err == nil {
		t.Fatal("Expected error when token not found in DB")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}

	if err.Error() != "refresh token not found or expired" {
		t.Errorf("Expected 'refresh token not found or expired' error, got %v", err)
	}
}

func TestRefreshTokenRevokeError(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	// Generate a valid refresh token
	refreshToken, _ := jwtManager.GenerateRefreshToken(1, "test@example.com")

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		GetByHashFunc: func(ctx context.Context, tokenHash string) (*repository.RefreshToken, error) {
			return &repository.RefreshToken{
				UserID:    1,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
		RevokeFunc: func(ctx context.Context, tokenHash string) error {
			return fmt.Errorf("revoke failed")
		},
	}

	mockUserRepo := &MockUserRepositoryAuth{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	result, err := service.RefreshToken(ctx, refreshToken)
	if err == nil {
		t.Fatal("Expected error when revoke fails")
	}

	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
}

// Tests for Logout

func TestLogoutSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		RevokeAllFunc: func(ctx context.Context, userID int) error {
			return nil
		},
	}

	mockUserRepo := &MockUserRepositoryAuth{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	err := service.Logout(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLogoutRevokeError(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		RevokeAllFunc: func(ctx context.Context, userID int) error {
			return fmt.Errorf("revoke all failed")
		},
	}

	mockUserRepo := &MockUserRepositoryAuth{}

	service := usecase.NewAuthService(mockUserRepo, mockRefreshTokenRepo, passwordManager, jwtManager)

	err := service.Logout(ctx, 1)
	if err == nil {
		t.Fatal("Expected error when revoke all fails")
	}

	if err.Error() != "failed to revoke refresh tokens: revoke all failed" {
		t.Errorf("Expected 'failed to revoke refresh tokens' error, got %v", err)
	}
}
