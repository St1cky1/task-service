package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// валидация
type CreateUserRequest struct {
	Name string `json:"name" validate:"required, min=1, max=255"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required, min=1, max=255"`
}
