package repo

import (
	"context"

	"github.com/St1cky1/task-service/internal/models"
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
func (r *UserRepository) Create(ctx context.Context, user *models.CreateUserRequest) (*models.User, error) {

	query := `
	INSERT INTO "user" (name)
	VALUES ($1)
	RETURNING id, name, created_at, updated_at
	`

	var createdUser models.User

	err := r.db.QueryRow(ctx, query, user.Name).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &createdUser, nil
}

// получаем данные по id
func (r *UserRepository) GetById(ctx context.Context, id int) (*models.User, error) {
	query := `
	SELECT id, name, created_at, updated_at 
	FROM "user"
	WHERE  id = ($1)
	`
	var user models.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
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
