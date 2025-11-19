package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}

type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

// Save - сохраняем refresh token
func (r *RefreshTokenRepository) Save(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	query := `
	INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return err
	}

	return nil
}

// GetByUserID - получаем невозвращенные токены пользователя
func (r *RefreshTokenRepository) GetByUserID(ctx context.Context, userID int) ([]RefreshToken, error) {
	query := `
	SELECT id, user_id, token_hash, expires_at, created_at, revoked
	FROM refresh_tokens
	WHERE user_id = $1 AND revoked = false AND expires_at > NOW()
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []RefreshToken
	for rows.Next() {
		var token RefreshToken
		err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.TokenHash,
			&token.ExpiresAt,
			&token.CreatedAt,
			&token.Revoked,
		)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

// RevokeAll - откатываем все токены пользователя
func (r *RefreshTokenRepository) RevokeAll(ctx context.Context, userID int) error {
	query := `
	UPDATE refresh_tokens
	SET revoked = true
	WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

// Revoke - откатываем конкретный токен
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	query := `
	UPDATE refresh_tokens
	SET revoked = true
	WHERE token_hash = $1
	`

	_, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}

	return nil
}

// GetByHash - получаем токен по хешу
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
	SELECT id, user_id, token_hash, expires_at, created_at, revoked
	FROM refresh_tokens
	WHERE token_hash = $1 AND revoked = false AND expires_at > NOW()
	`

	var token RefreshToken
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.Revoked,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &token, nil
}

// CleanupExpired - удаляем истекшие токены
func (r *RefreshTokenRepository) CleanupExpired(ctx context.Context) error {
	query := `
	DELETE FROM refresh_tokens
	WHERE expires_at < NOW()
	`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
