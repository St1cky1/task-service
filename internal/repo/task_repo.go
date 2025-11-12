package repo

import (
	"context"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *models.CreateTaskRequest) (*models.Task, error) {

	query := `
	INSERT INTO "task" (title, description, status, owner_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, title, description, status, owner_id, created_at, updated_at
	`

	var createdTask models.Task
	err := r.db.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.OwnerId,
	).Scan(
		&createdTask.Title,
		&createdTask.Description,
		&createdTask.Status,
		&createdTask.OwnerId,
		&createdTask.CreatedAt,
		&createdTask.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &createdTask, nil
}

func (r *TaskRepository) GetByTaskId(ctx context.Context, taskId int) (*models.Task, error) {

	query := `
	SELECT id, title, description, owner_id, created_at, updated_at
	FROM "task"
	WHERE id = &1
	`
	var task models.Task

	err := r.db.QueryRow(ctx, query, taskId).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.OwnerId,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}
