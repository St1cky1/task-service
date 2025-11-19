package entity

import "time"

type User struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Email        *string    `json:"email,omitempty"`
	PasswordHash string     `json:"-"` // Никогда не отправляем пароль
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// валидация
type CreateUserRequest struct {
	Name string `json:"name" validate:"required, min=1, max=255"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required, min=1, max=255"`
}

// Регистрация
type RegisterRequest struct {
	Name     string `json:"name" validate:"required, min=1, max=255"`
	Email    string `json:"email" validate:"required, email"`
	Password string `json:"password" validate:"required, min=8, max=255"`
}

// Логин
type LoginRequest struct {
	Email    string `json:"email" validate:"required, email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Refresh Token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JWT Claims
type JWTClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}
