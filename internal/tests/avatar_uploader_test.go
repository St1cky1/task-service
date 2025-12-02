package usecase_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/infrastructure/auth"
	"github.com/St1cky1/task-service/internal/usecase"
)

// Tests for generateRandomPassword

func TestGenerateRandomPassword(t *testing.T) {
	password := usecase.GenerateRandomPassword(12)

	if len(password) != 12 {
		t.Errorf("Expected password length 12, got %d", len(password))
	}

	if password == "" {
		t.Error("Expected non-empty password")
	}
}

func TestGenerateRandomPasswordDifferent(t *testing.T) {
	password1 := usecase.GenerateRandomPassword(16)
	password2 := usecase.GenerateRandomPassword(16)

	if password1 == password2 {
		t.Error("Expected different passwords")
	}
}

func TestGenerateRandomPasswordLength(t *testing.T) {
	lengths := []int{8, 12, 20, 32}

	for _, length := range lengths {
		password := usecase.GenerateRandomPassword(length)
		if len(password) != length {
			t.Errorf("Expected password length %d, got %d", length, len(password))
		}
	}
}

// Tests for uploadUserAvatar

func TestUploadUserAvatarSuccess(t *testing.T) {
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	imageData := []byte("test image data for avatar")
	err := usecase.UploadUserAvatar(ctx, userService, 1, imageData)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUploadUserAvatarUserNotFound(t *testing.T) {
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	imageData := []byte("test image data")
	err := usecase.UploadUserAvatar(ctx, userService, 999, imageData)

	if err == nil {
		t.Fatal("Expected error for user not found")
	}
}

func TestUploadUserAvatarFileSizeExceedsLimit(t *testing.T) {
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	// Create data larger than 5MB
	largeData := make([]byte, 6*1024*1024)
	err := usecase.UploadUserAvatar(ctx, userService, 1, largeData)

	if err == nil {
		t.Fatal("Expected error for file size exceeding limit")
	}
}

// Tests for DownloadAvatarStream

func TestDownloadAvatarStreamSuccess(t *testing.T) {
	ctx := context.Background()
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	testData := []byte("test image data for streaming")

	// Create a temporary file
	tmpFile := "/tmp/test_avatar_stream"
	err := os.WriteFile(tmpFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	dataChan, errChan := userService.DownloadAvatarStream(ctx, 1, 1024)

	var receivedData []byte
	for data := range dataChan {
		receivedData = append(receivedData, data...)
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	default:
		// No error channel message is expected on success
	}

	if string(receivedData) != string(testData) {
		t.Errorf("Expected data %s, got %s", string(testData), string(receivedData))
	}
}

func TestDownloadAvatarStreamUserNotFound(t *testing.T) {
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	dataChan, errChan := userService.DownloadAvatarStream(ctx, 999, 1024)

	// Consume the channels
	for range dataChan {
	}

	var errorReceived bool
	select {
	case err := <-errChan:
		if err != nil {
			errorReceived = true
		}
	case <-time.After(1 * time.Second):
		// Timeout
	}

	if !errorReceived {
		t.Error("Expected error for user not found")
	}
}

func TestDownloadAvatarStreamAvatarNotFound(t *testing.T) {
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	dataChan, errChan := userService.DownloadAvatarStream(ctx, 1, 1024)

	// Consume the channels
	for range dataChan {
	}

	var errorReceived bool
	select {
	case err := <-errChan:
		if err != nil && fmt.Sprintf("%v", err) == "avatar not found" {
			errorReceived = true
		}
	case <-time.After(1 * time.Second):
		// Timeout
	}

	if !errorReceived {
		t.Error("Expected error for avatar not found")
	}
}

func TestDownloadAvatarStreamContextCancellation(t *testing.T) {
	passwordManager := auth.NewPasswordManager()
	jwtManager := auth.NewJWTManager()

	mockUser := &entity.User{
		ID:   1,
		Name: "Test User",
	}

	testData := make([]byte, 10*1024) // 10KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Create a temporary file
	tmpFile := "/tmp/test_avatar_stream_cancel"
	err := os.WriteFile(tmpFile, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
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

	userService := usecase.NewUserService(mockUserRepo, mockAvatarRepo, passwordManager, jwtManager, mockRefreshTokenRepo)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	dataChan, errChan := userService.DownloadAvatarStream(ctx, 1, 512)

	// Receive one chunk and then cancel
	<-dataChan
	cancel()

	// Try to receive more data
	var receivedData bool
	select {
	case <-dataChan:
		receivedData = true
	case <-time.After(1 * time.Second):
		// Expected timeout or channel close
	}

	// Check if context cancellation error was received
	var cancelError bool
	select {
	case err := <-errChan:
		if err != nil {
			cancelError = true
		}
	case <-time.After(1 * time.Second):
		// Channel might close before error is sent
	}

	// Either the channel should close or we should get a context error
	if receivedData && !cancelError {
		t.Error("Expected context cancellation")
	}
}

// Benchmarks

func BenchmarkGenerateRandomPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		usecase.GenerateRandomPassword(12)
	}
}

func BenchmarkGenerateRandomPasswordLonger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		usecase.GenerateRandomPassword(32)
	}
}
