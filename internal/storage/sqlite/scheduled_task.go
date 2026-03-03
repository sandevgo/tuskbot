package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ScheduledTaskRepo struct {
	db *sql.DB
}

func NewScheduledTaskRepo(db *sql.DB) *ScheduledTaskRepo {
	return &ScheduledTaskRepo{db: db}
}

func (r *ScheduledTaskRepo) Create(ctx context.Context, task core.StoredTask) error {
	query := `
		INSERT INTO task (id, name, owner_session_id, prompt, trigger_type, trigger_spec, last_run, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.Name,
		task.OwnerSessionID,
		task.Prompt,
		task.TriggerType,
		task.TriggerSpec,
		task.LastRun,
		task.IsActive,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create scheduled task: %w", err)
	}
	return nil
}

func (r *ScheduledTaskRepo) Cancel(ctx context.Context, name string) error {
	query := `
		UPDATE task 
		SET is_active = FALSE, updated_at = ?
		WHERE name = ? AND is_active = TRUE
	`
	result, err := r.db.ExecContext(ctx, query, time.Now(), name)
	if err != nil {
		return fmt.Errorf("failed to cancel scheduled task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found or already cancelled: %s", name)
	}

	return nil
}

func (r *ScheduledTaskRepo) List(ctx context.Context) ([]core.StoredTask, error) {
	query := `
		SELECT id, name, owner_session_id, prompt, trigger_type, trigger_spec, last_run, is_active, created_at, updated_at
		FROM task
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled tasks: %w", err)
	}
	defer rows.Close()

	var tasks []core.StoredTask
	for rows.Next() {
		var task core.StoredTask
		var lastRun sql.NullTime
		var updatedAt sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.OwnerSessionID,
			&task.Prompt,
			&task.TriggerType,
			&task.TriggerSpec,
			&lastRun,
			&task.IsActive,
			&task.CreatedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduled task: %w", err)
		}

		if lastRun.Valid {
			task.LastRun = lastRun.Time
		}
		if updatedAt.Valid {
			task.UpdatedAt = &updatedAt.Time
		}

		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}
