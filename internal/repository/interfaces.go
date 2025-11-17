package repository

import (
	"context"

	"github.com/St1cky1/task-service/internal/entity"
)

// ITaskRepository - интерфейс для TaskRepository
type ITaskRepository interface {
	Create(ctx context.Context, task *entity.CreateTaskRequest) (*entity.Task, error)
	GetByTaskId(ctx context.Context, taskId int) (*entity.Task, error)
	Update(ctx context.Context, id int, updates map[string]interface{}) (*entity.Task, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, ownerID int, status string) ([]entity.Task, error)
}

// IUserRepository - интерфейс для UserRepository
type IUserRepository interface {
	Create(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error)
	GetById(ctx context.Context, id int) (*entity.User, error)
	Update(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error)
	List(ctx context.Context) ([]entity.User, error)
	Delete(ctx context.Context, id int) error
}

// IAvatarRepository - интерфейс для AvatarRepository
type IAvatarRepository interface {
	Save(ctx context.Context, avatar *entity.Avatar) (*entity.Avatar, error)
	GetByUserId(ctx context.Context, userId int) (*entity.Avatar, error)
	DeleteByUserId(ctx context.Context, userId int) error
}

// ITaskAuditRepository - интерфейс для TaskAuditRepository
type ITaskAuditRepository interface {
	Create(ctx context.Context, audit *entity.TaskAudit) error
	GetByTaskAuditId(ctx context.Context, taskAuditId int) ([]entity.TaskAudit, error)
}
