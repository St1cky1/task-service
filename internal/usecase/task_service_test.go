package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/St1cky1/task-service/internal/repository"
)

// MockTaskRepository - мок для ITaskRepository
type MockTaskRepository struct {
	CreateFunc      func(ctx context.Context, task *entity.CreateTaskRequest) (*entity.Task, error)
	GetByTaskIdFunc func(ctx context.Context, taskId int) (*entity.Task, error)
	UpdateFunc      func(ctx context.Context, id int, updates map[string]interface{}) (*entity.Task, error)
	DeleteFunc      func(ctx context.Context, id int) error
	ListFunc        func(ctx context.Context, ownerID int, status string) ([]entity.Task, error)
}

var _ repository.ITaskRepository = (*MockTaskRepository)(nil)

func (m *MockTaskRepository) Create(ctx context.Context, task *entity.CreateTaskRequest) (*entity.Task, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, task)
	}
	return nil, nil
}

func (m *MockTaskRepository) GetByTaskId(ctx context.Context, taskId int) (*entity.Task, error) {
	if m.GetByTaskIdFunc != nil {
		return m.GetByTaskIdFunc(ctx, taskId)
	}
	return nil, nil
}

func (m *MockTaskRepository) Update(ctx context.Context, id int, updates map[string]interface{}) (*entity.Task, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates)
	}
	return nil, nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockTaskRepository) List(ctx context.Context, ownerID int, status string) ([]entity.Task, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, ownerID, status)
	}
	return nil, nil
}

// MockUserRepository - мок для IUserRepository
type MockUserRepository struct {
	GetByIdFunc func(ctx context.Context, id int) (*entity.User, error)
	CreateFunc  func(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error)
}

var _ repository.IUserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) GetById(ctx context.Context, id int) (*entity.User, error) {
	if m.GetByIdFunc != nil {
		return m.GetByIdFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil, nil
}

// MockTaskAuditRepository - мок для ITaskAuditRepository
type MockTaskAuditRepository struct {
	CreateFunc           func(ctx context.Context, audit *entity.TaskAudit) error
	GetByTaskAuditIdFunc func(ctx context.Context, taskAuditId int) ([]entity.TaskAudit, error)
}

var _ repository.ITaskAuditRepository = (*MockTaskAuditRepository)(nil)

func (m *MockTaskAuditRepository) Create(ctx context.Context, audit *entity.TaskAudit) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, audit)
	}
	return nil
}

func (m *MockTaskAuditRepository) GetByTaskAuditId(ctx context.Context, taskAuditId int) ([]entity.TaskAudit, error) {
	if m.GetByTaskAuditIdFunc != nil {
		return m.GetByTaskAuditIdFunc(ctx, taskAuditId)
	}
	return nil, nil
}

// MockRabbitMQPublisher - мок для RabbitMQPublisher
type MockRabbitMQPublisher struct {
	PublishAuditMessageFunc func(ctx context.Context, message *entity.AuditMessage) error
}

func (m *MockRabbitMQPublisher) PublishAuditMessage(ctx context.Context, message *entity.AuditMessage) error {
	if m.PublishAuditMessageFunc != nil {
		return m.PublishAuditMessageFunc(ctx, message)
	}
	return nil
}

// Tests

func TestCreateTaskSuccess(t *testing.T) {
	ctx := context.Background()
	mockUser := &entity.User{ID: 1, Name: "Test User"}
	mockTask := &entity.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.StatusPending,
		OwnerId:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUserRepo := &MockUserRepository{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			if id == 1 {
				return mockUser, nil
			}
			return nil, nil
		},
	}

	mockTaskRepo := &MockTaskRepository{
		CreateFunc: func(ctx context.Context, task *entity.CreateTaskRequest) (*entity.Task, error) {
			return mockTask, nil
		},
	}

	mockAuditRepo := &MockTaskAuditRepository{}
	mockRabbitMQ := &MockRabbitMQPublisher{}

	service := NewTaskService(mockTaskRepo, mockUserRepo, mockAuditRepo, mockRabbitMQ)

	req := &entity.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.StatusPending,
	}

	result, err := service.CreateTask(ctx, req, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != mockTask.ID {
		t.Errorf("Expected task ID %d, got %d", mockTask.ID, result.ID)
	}

	if result.Title != mockTask.Title {
		t.Errorf("Expected title %s, got %s", mockTask.Title, result.Title)
	}
}

func TestCreateTaskUserNotFound(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := &MockUserRepository{
		GetByIdFunc: func(ctx context.Context, id int) (*entity.User, error) {
			return nil, nil // User not found
		},
	}

	mockTaskRepo := &MockTaskRepository{}
	mockAuditRepo := &MockTaskAuditRepository{}
	mockRabbitMQ := &MockRabbitMQPublisher{}

	service := NewTaskService(mockTaskRepo, mockUserRepo, mockAuditRepo, mockRabbitMQ)

	req := &entity.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entity.StatusPending,
	}

	result, err := service.CreateTask(ctx, req, 999)
	if err != entity.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil task, got %v", result)
	}
}

func TestUpdateTaskSuccess(t *testing.T) {
	ctx := context.Background()
	oldTask := &entity.Task{
		ID:          1,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      entity.StatusPending,
		OwnerId:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updatedTask := &entity.Task{
		ID:          1,
		Title:       "New Title",
		Description: "Old Description",
		Status:      entity.StatusCompleted,
		OwnerId:     1,
		CreatedAt:   oldTask.CreatedAt,
		UpdatedAt:   time.Now(),
	}

	mockTaskRepo := &MockTaskRepository{
		GetByTaskIdFunc: func(ctx context.Context, taskId int) (*entity.Task, error) {
			if taskId == 1 {
				return oldTask, nil
			}
			return nil, nil
		},
		UpdateFunc: func(ctx context.Context, id int, updates map[string]interface{}) (*entity.Task, error) {
			return updatedTask, nil
		},
	}

	mockUserRepo := &MockUserRepository{}
	mockAuditRepo := &MockTaskAuditRepository{}
	mockRabbitMQ := &MockRabbitMQPublisher{}

	service := NewTaskService(mockTaskRepo, mockUserRepo, mockAuditRepo, mockRabbitMQ)

	req := &entity.UpdateTaskRequest{
		Title:  "New Title",
		Status: entity.StatusCompleted,
	}

	result, err := service.UpdateTask(ctx, 1, 1, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Title != updatedTask.Title {
		t.Errorf("Expected title %s, got %s", updatedTask.Title, result.Title)
	}

	if result.Status != updatedTask.Status {
		t.Errorf("Expected status %s, got %s", updatedTask.Status, result.Status)
	}
}

func TestUpdateTaskNotFound(t *testing.T) {
	ctx := context.Background()

	mockTaskRepo := &MockTaskRepository{
		GetByTaskIdFunc: func(ctx context.Context, taskId int) (*entity.Task, error) {
			return nil, nil // Task not found
		},
	}

	mockUserRepo := &MockUserRepository{}
	mockAuditRepo := &MockTaskAuditRepository{}
	mockRabbitMQ := &MockRabbitMQPublisher{}

	service := NewTaskService(mockTaskRepo, mockUserRepo, mockAuditRepo, mockRabbitMQ)

	req := &entity.UpdateTaskRequest{
		Title: "New Title",
	}

	result, err := service.UpdateTask(ctx, 999, 1, req)
	if err != entity.ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}

	if result != nil {
		t.Errorf("Expected nil task, got %v", result)
	}
}
