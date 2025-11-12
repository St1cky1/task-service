package models

import "time"

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusInProgres TaskStatus = "in_progres"
	StatusCompleted TaskStatus = "completed"
	StatusCancelled TaskStatus = "cancelled"
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
	Status      TaskStatus `json:"status" validate:"oneof=pending in_progres completed cancelled"`
	OwnerId     int        `json:"owner_id" validate:"required, min=1"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title" validate:"required, min=1, max=255"`
	Description *string    `json:"description" validate:"required"` // проверяем пустая строка или не пердана
	Status      TaskStatus `json:"status" validate:"oneof=pending in_progres completed cancelled"`
}
