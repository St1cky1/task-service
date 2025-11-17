package repository

import (
	"context"
	"strconv"

	"github.com/St1cky1/task-service/internal/entity"
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

func (r *TaskRepository) Create(ctx context.Context, task *entity.CreateTaskRequest) (*entity.Task, error) {

	query := `
	INSERT INTO "task" (title, description, status, owner_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, title, description, status, owner_id, created_at, updated_at
	`

	var createdTask entity.Task
	err := r.db.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.OwnerId,
	).Scan(
		&createdTask.ID,
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

func (r *TaskRepository) GetByTaskId(ctx context.Context, taskId int) (*entity.Task, error) {

	query := `
	SELECT id, title, description, status, owner_id, created_at, updated_at
	FROM "task"
	WHERE id = $1
	`
	var task entity.Task

	err := r.db.QueryRow(ctx, query, taskId).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
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

// Update - обновление задачи
func (r *TaskRepository) Update(ctx context.Context, id int, updates map[string]interface{}) (*entity.Task, error) {
	// Динамически строим SET часть запроса
	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		if field == "updated_at" {
			continue // не обновляем вручную
		}
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += field + " = $" + strconv.Itoa(argIndex)
		args = append(args, value)
		argIndex++
	}

	// Добавляем обновление updated_at
	if argIndex > 1 {
		setClause += ", updated_at = CURRENT_TIMESTAMP"
	}

	query := `
        UPDATE task 
        SET ` + setClause + `
        WHERE id = $` + strconv.Itoa(argIndex) + `
        RETURNING id, title, description, status, owner_id, created_at, updated_at
    `
	args = append(args, id)

	var task entity.Task
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.OwnerId,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

// Delete - удаление задачи
func (r *TaskRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM task WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// List - список задач с фильтрацией
func (r *TaskRepository) List(ctx context.Context, ownerID int, status string) ([]entity.Task, error) {
	query := `
        SELECT id, title, description, status, owner_id, created_at, updated_at 
        FROM task 
        WHERE owner_id = $1
    `
	args := []interface{}{ownerID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []entity.Task
	for rows.Next() {
		var task entity.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.OwnerId,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
