package repository

import (
	"context"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// создаем пользователя
func (r *UserRepository) Create(ctx context.Context, user *entity.CreateUserRequest) (*entity.User, error) {

	query := `
	INSERT INTO "user" (name)
	VALUES ($1)
	RETURNING id, name, email, password_hash, avatar_url, is_active, last_login, created_at, updated_at
	`

	var createdUser entity.User

	err := r.db.QueryRow(ctx, query, user.Name).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.PasswordHash,
		&createdUser.AvatarURL,
		&createdUser.IsActive,
		&createdUser.LastLogin,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdUser, nil
}

// получаем данные по id
func (r *UserRepository) GetById(ctx context.Context, id int) (*entity.User, error) {
	query := `
	SELECT id, name, email, password_hash, avatar_url, is_active, last_login, created_at, updated_at 
	FROM "user"
	WHERE  id = ($1)
	`
	var user entity.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// Update - обновляем пользователя
func (r *UserRepository) Update(ctx context.Context, id int, updates map[string]interface{}) (*entity.User, error) {
	query := `
	UPDATE "user"
	SET name = COALESCE($1, name),
	    email = COALESCE($2, email),
	    avatar_url = COALESCE($3, avatar_url),
	    last_login = COALESCE($4, last_login),
	    updated_at = CURRENT_TIMESTAMP
	WHERE id = $5
	RETURNING id, name, email, password_hash, avatar_url, is_active, last_login, created_at, updated_at
	`

	var user entity.User

	var name interface{} = updates["name"]
	var email interface{} = updates["email"]
	var avatarURL interface{} = updates["avatar_url"]
	var lastLogin interface{} = updates["last_login"]

	err := r.db.QueryRow(ctx, query, name, email, avatarURL, lastLogin, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// List - получаем всех пользователей
func (r *UserRepository) List(ctx context.Context) ([]entity.User, error) {
	query := `
	SELECT id, name, email, password_hash, avatar_url, is_active, last_login, created_at, updated_at 
	FROM "user"
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Delete - удаляем пользователя
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM "user" WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// GetByEmail - получаем пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
	SELECT id, name, email, password_hash, avatar_url, is_active, last_login, created_at, updated_at 
	FROM "user"
	WHERE email = $1
	`
	var user entity.User

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// CreateWithAuth - создаем пользователя с email и паролем
func (r *UserRepository) CreateWithAuth(ctx context.Context, name, email, passwordHash string) (*entity.User, error) {
	query := `
	INSERT INTO "user" (name, email, password_hash, is_active)
	VALUES ($1, $2, $3, true)
	RETURNING id, name, email, password_hash, avatar_url, is_active, last_login, created_at, updated_at
	`

	var user entity.User

	err := r.db.QueryRow(ctx, query, name, email, passwordHash).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
