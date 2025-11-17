package repository

import (
	"context"

	"github.com/St1cky1/task-service/internal/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AvatarRepository struct {
	db *pgxpool.Pool
}

func NewAvatarRepository(db *pgxpool.Pool) *AvatarRepository {
	return &AvatarRepository{
		db: db,
	}
}

// Save - создает или обновляет аватарку
func (r *AvatarRepository) Save(ctx context.Context, avatar *entity.Avatar) (*entity.Avatar, error) {
	query := `
	INSERT INTO "avatar" (user_id, file_path, file_size, content_type)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE SET
	    file_path = $2,
	    file_size = $3,
	    content_type = $4,
	    updated_at = CURRENT_TIMESTAMP
	RETURNING id, user_id, file_path, file_size, content_type, created_at, updated_at
	`

	var savedAvatar entity.Avatar

	err := r.db.QueryRow(ctx, query, avatar.UserID, avatar.FilePath, avatar.FileSize, avatar.ContentType).Scan(
		&savedAvatar.ID,
		&savedAvatar.UserID,
		&savedAvatar.FilePath,
		&savedAvatar.FileSize,
		&savedAvatar.ContentType,
		&savedAvatar.CreatedAt,
		&savedAvatar.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &savedAvatar, nil
}

// GetByUserId - получает аватарку по user_id
func (r *AvatarRepository) GetByUserId(ctx context.Context, userId int) (*entity.Avatar, error) {
	query := `
	SELECT id, user_id, file_path, file_size, content_type, created_at, updated_at
	FROM "avatar"
	WHERE user_id = $1
	`

	var avatar entity.Avatar

	err := r.db.QueryRow(ctx, query, userId).Scan(
		&avatar.ID,
		&avatar.UserID,
		&avatar.FilePath,
		&avatar.FileSize,
		&avatar.ContentType,
		&avatar.CreatedAt,
		&avatar.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &avatar, nil
}

// DeleteByUserId - удаляет аватарку по user_id
func (r *AvatarRepository) DeleteByUserId(ctx context.Context, userId int) error {
	query := `DELETE FROM "avatar" WHERE user_id = $1`
	result, err := r.db.Exec(ctx, query, userId)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
