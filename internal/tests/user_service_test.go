package usecase_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/infrastructure/auth"
	"github.com/St1cky1/task-service/internal/repository"
	"github.com/St1cky1/task-service/internal/usecase"
)

// MockAvatarRepository для тестирования UserService
type MockAvatarRepository struct {
	SaveFunc           func(ctx context.Context, avatar *entity.Avatar) (*entity.Avatar, error)
	GetByUserIdFunc    func(ctx context.Context, userId int) (*entity.Avatar, error)
	DeleteByUserIdFunc func(ctx context.Context, userId int) error
}

var _ repository.IAvatarRepository = (*MockAvatarRepository)(nil)

func (m *MockAvatarRepository) Save(ctx context.Context, avatar *entity.Avatar) (*entity.Avatar, error) {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, avatar)
	}
	return nil, nil
}

func (m *MockAvatarRepository) GetByUserId(ctx context.Context, userId int) (*entity.Avatar, error) {
	if m.GetByUserIdFunc != nil {
		return m.GetByUserIdFunc(ctx, userId)
	}
	return nil, nil
}

func (m *MockAvatarRepository) DeleteByUserId(ctx context.Context, userId int) error {
	if m.DeleteByUserIdFunc != nil {
		return m.DeleteByUserIdFunc(ctx, userId)
	}
	return nil
}

// Tests for CreateUser

func TestCreateUserSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:        1,
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUserRepo := &MockUserRepositoryAuth{
		CreateFunc: func(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	req := &entity.CreateUserRequest{
		Name: "Test User",
	}

	result, err := service.CreateUser(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != mockUser.ID {
		t.Errorf("Expected user ID %d, got %d", mockUser.ID, result.ID)
	}

	if result.Name != mockUser.Name {
		t.Errorf("Expected name %s, got %s", mockUser.Name, result.Name)
	}
}

func TestCreateUserError(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		CreateFunc: func(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	req := &entity.CreateUserRequest{
		Name: "Test User",
	}

	result, err := service.CreateUser(ctx, req)
	if err == nil {
		t.Fatal("Expected error")
	}

	if result != nil {
		t.Errorf("Expected nil result")
	}
}

// Tests for GetUser

func TestGetUserSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:       1,
		Name:     "Test User",
		IsActive: true,
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	result, err := service.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != mockUser.ID {
		t.Errorf("Expected user ID %d, got %d", mockUser.ID, result.ID)
	}
}

func TestGetUserNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	result, err := service.GetUser(ctx, 999)
	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result")
	}
}

// Tests for UpdateUser

func TestUpdateUserSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "New Name",
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			if id == 1 {
				return &entity.User{ID: 1, Name: "Old Name"}, nil
			}
			return nil, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	req := &entity.UpdateUserRequest{
		Name: "New Name",
	}

	result, err := service.UpdateUser(ctx, 1, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "New Name" {
		t.Errorf("Expected name 'New Name', got %s", result.Name)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	req := &entity.UpdateUserRequest{
		Name: "New Name",
	}

	result, err := service.UpdateUser(ctx, 999, req)
	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil result")
	}
}

// Tests for DeleteUser

func TestDeleteUserSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
		DeleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return nil, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	err := service.DeleteUser(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteUserWithAvatar(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	mockAvatar := &entity.Avatar{
		ID:       1,
		UserID:   1,
		FilePath: "/tmp/test_avatar",
	}

	// Create a temporary file
	tmpFile := "/tmp/test_avatar"
	os.WriteFile(tmpFile, []byte("test"), 0644)
	defer os.Remove(tmpFile)

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
		DeleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			if userId == 1 {
				return mockAvatar, nil
			}
			return nil, nil
		},
		DeleteByUserIdFunc: func(ctx context.Context, userId int) error {
			return nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	err := service.DeleteUser(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	err := service.DeleteUser(ctx, 999)
	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

// Tests for ListUsers

func TestListUsersSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	users := []entity.User{
		{ID: 1, Name: "User 1"},
		{ID: 2, Name: "User 2"},
	}

	mockUserRepo := &MockUserRepositoryAuth{
		ListFunc: func(ctx context.Context) ([]entity.User, error) {
			return users, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	result, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 users, got %d", len(result))
	}
}

func TestListUsersEmpty(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		ListFunc: func(ctx context.Context) ([]entity.User, error) {
			return []entity.User{}, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	result, err := service.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 users, got %d", len(result))
	}
}

// Tests for UploadAvatar

func TestUploadAvatarSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return nil, nil
		},
		SaveFunc: func(ctx context.Context, avatar *entity.Avatar) (*entity.Avatar, error) {
			return avatar, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	imageData := []byte("test image data")
	result, err := service.UploadAvatar(ctx, 1, imageData, "image/jpeg")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty file path")
	}

	// Clean up
	if result != "" {
		os.RemoveAll("var/avatars")
	}
}

func TestUploadAvatarUserNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	imageData := []byte("test image data")
	result, err := service.UploadAvatar(ctx, 999, imageData, "image/jpeg")

	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty file path")
	}
}

func TestUploadAvatarFileSizeTooLarge(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	// Create data larger than 5MB
	largeData := make([]byte, 6*1024*1024)
	result, err := service.UploadAvatar(ctx, 1, largeData, "image/jpeg")

	if err == nil {
		t.Fatal("Expected error for file size exceeding limit")
	}

	if result != "" {
		t.Errorf("Expected empty file path")
	}
}

// Tests for DownloadAvatar

func TestDownloadAvatarSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	testData := []byte("test image data")

	// Create a temporary file
	tmpFile := "/tmp/test_avatar_download"
	os.WriteFile(tmpFile, testData, 0644)
	defer os.Remove(tmpFile)

	mockAvatar := &entity.Avatar{
		ID:          1,
		UserID:      1,
		FilePath:    tmpFile,
		ContentType: "image/jpeg",
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return mockAvatar, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	data, contentType, err := service.DownloadAvatar(ctx, 1)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected data %s, got %s", string(testData), string(data))
	}

	if contentType != "image/jpeg" {
		t.Errorf("Expected content type 'image/jpeg', got %s", contentType)
	}
}

func TestDownloadAvatarUserNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	data, contentType, err := service.DownloadAvatar(ctx, 999)

	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty data")
	}

	if contentType != "" {
		t.Errorf("Expected empty content type")
	}
}

func TestDownloadAvatarNotFound(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	mockUserRepo := &MockUserRepositoryAuth{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return nil, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	data, contentType, err := service.DownloadAvatar(ctx, 1)

	if err == nil {
		t.Fatal("Expected error for avatar not found")
	}

	if len(data) != 0 {
		t.Errorf("Expected empty data")
	}

	if contentType != "" {
		t.Errorf("Expected empty content type")
	}
}

// Tests for HasAvatar

func TestHasAvatarTrue(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockAvatar := &entity.Avatar{
		ID:       1,
		UserID:   1,
		FilePath: "/tmp/avatar",
	}

	mockUserRepo := &MockUserRepositoryAuth{}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return mockAvatar, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	result := service.HasAvatar(ctx, 1)
	if !result {
		t.Error("Expected HasAvatar to return true")
	}
}

func TestHasAvatarFalse(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return nil, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	result := service.HasAvatar(ctx, 1)
	if result {
		t.Error("Expected HasAvatar to return false")
	}
}

// Tests for CreateUserWithAvatar

func TestCreateUserWithAvatarSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:        1,
		Name:      "Test User",
		Email:     func() *string { e := "test@example.com"; return &e }(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUserRepo := &MockUserRepositoryAuth{
		CreateWithAuthFunc: func(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
			return mockUser, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
			return mockUser, nil
		},
	}

	mockAvatarRepo := &MockAvatarRepository{
		GetByUserIdFunc: func(ctx context.Context, userId int) (*entity.Avatar, error) {
			return nil, nil
		},
		SaveFunc: func(ctx context.Context, avatar *entity.Avatar) (*entity.Avatar, error) {
			return avatar, nil
		},
	}

	mockRefreshTokenRepo := &MockRefreshTokenRepository{
		SaveFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	imageData := []byte("test image data")
	result, err := service.CreateUserWithAvatar(ctx, "Test User", "test@example.com", "password123", imageData)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != mockUser.ID {
		t.Errorf("Expected user ID %d, got %d", mockUser.ID, result.ID)
	}

	// Clean up
	os.RemoveAll("var/avatars")
}

func TestCreateUserWithAvatarCreateWithAuthError(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUserRepo := &MockUserRepositoryAuth{
		CreateWithAuthFunc: func(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	mockAvatarRepo := &MockAvatarRepository{}
	mockRefreshTokenRepo := &MockRefreshTokenRepository{}

	service := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	imageData := []byte("test image data")
	result, err := service.CreateUserWithAvatar(ctx, "Test User", "test@example.com", "password123", imageData)

	if err == nil {
		t.Fatal("Expected error when CreateWithAuth fails")
	}

	if result != nil {
		t.Errorf("Expected nil result")
	}
}
