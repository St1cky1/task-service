package repo

import (
	"context"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskAuditRepository struct {
	db *pgxpool.Pool
}

func NewTaskAuditRepository(db *pgxpool.Pool) *TaskAuditRepository {
	return &TaskAuditRepository{
		db: db,
	}
}

func (r *TaskAuditRepository) Create(ctx context.Context, audit *models.TaskAudit) error {
	query := `
	INSERT INTO "task_audit" (user_id, actions, entity_type, entity_id, old_values, new_values, changes)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	RETURNING id, changed_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		audit.UserID,
		audit.Action,
		audit.EntityType,
		audit.EntityID,
		audit.OldValues,
		audit.NewValues,
		audit.Changes,
	).Scan(&audit.ID, &audit.ChangesAt)

	return err
}

func (r *TaskAuditRepository) GetByTaskAuditId(ctx context.Context, taskAuditId int) ([]models.TaskAudit, error) {
	query := `
	SELECT id, user_id, action, entity_type, entity_id, old_values, new_values, changes, changed_at
	FROM "task_audit"
	WHERE entity_id = $1 and entity_type = 'task'
	ORDER BY changed_at DESC
	`
	rows, err := r.db.Query(ctx, query, taskAuditId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []models.TaskAudit
	for rows.Next() {
		var audit models.TaskAudit
		err := rows.Scan(
			&audit.ID,
			&audit.UserID,
			&audit.Action,
			&audit.EntityType,
			&audit.EntityID,
			&audit.OldValues,
			&audit.NewValues,
			&audit.Changes,
			&audit.ChangesAt,
		)
		if err != nil {
			return nil, err
		}
		audits = append(audits, audit)
	}
	return audits, nil
}
