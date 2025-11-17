package entity

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusCancelled  TaskStatus = "cancelled"
)

type Task struct {
	ID          int        `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	OwnerId     int        `json:"owner_id"`
}

// валидация
type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required, min=1, max=255"`
	Description string     `json:"description" validate:"required"`
	Status      TaskStatus `json:"status" validate:"oneof=pending in_progress completed cancelled"`
	OwnerId     int        `json:"owner_id" validate:"required, min=1"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description"` // опциональное поле для обновления
	Status      TaskStatus `json:"status"`
}
