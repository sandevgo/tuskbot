package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create table schema based on migration
	schema := `
	CREATE TABLE IF NOT EXISTS task (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		owner_session_id TEXT NOT NULL,
		prompt TEXT NOT NULL,
		trigger_type TEXT NOT NULL CHECK(trigger_type IN ('cron', 'once', 'interval')),
		trigger_spec TEXT NOT NULL,
		last_run DATETIME,
		is_active BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME
	);

	CREATE INDEX idx_task_owner ON task(owner_session_id);
	CREATE INDEX idx_task_active ON task(is_active) WHERE is_active = TRUE;
	`

	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestScheduledTaskRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTaskRepo(db)
	ctx := context.Background()

	t.Run("successfully creates task", func(t *testing.T) {
		task := &core.StoredTask{
			ID:             uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:           "test-task",
			OwnerSessionID: "session-123",
			Prompt:         "Test prompt",
			TriggerType:    "interval",
			TriggerSpec:    "1h",
			IsActive:       true,
			CreatedAt:      time.Now(),
		}

		err := repo.Create(ctx, task)
		require.NoError(t, err)

		// Verify task was created
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM task WHERE id = ?", task.ID.String()).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("fails on duplicate name", func(t *testing.T) {
		task := &core.StoredTask{
			ID:             uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			Name:           "duplicate-task",
			OwnerSessionID: "session-123",
			Prompt:         "Test prompt",
			TriggerType:    "cron",
			TriggerSpec:    "0 0 * * *",
			IsActive:       true,
			CreatedAt:      time.Now(),
		}

		err := repo.Create(ctx, task)
		require.NoError(t, err)

		// Try to create another task with same name
		task2 := &core.StoredTask{
			ID:             uuid.MustParse("33333333-3333-3333-3333-333333333333"),
			Name:           "duplicate-task",
			OwnerSessionID: "session-456",
			Prompt:         "Another prompt",
			TriggerType:    "once",
			TriggerSpec:    "2025-01-01T00:00:00Z",
			IsActive:       true,
			CreatedAt:      time.Now(),
		}

		err = repo.Create(ctx, task2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create scheduled task")
	})
}

func TestScheduledTaskRepo_Cancel(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTaskRepo(db)
	ctx := context.Background()

	t.Run("successfully cancels active task", func(t *testing.T) {
		task := &core.StoredTask{
			ID:             uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			Name:           "cancelable-task",
			OwnerSessionID: "session-123",
			Prompt:         "Test prompt",
			TriggerType:    "interval",
			TriggerSpec:    "1h",
			IsActive:       true,
			CreatedAt:      time.Now(),
		}

		err := repo.Create(ctx, task)
		require.NoError(t, err)

		err = repo.Cancel(ctx, task.Name)
		require.NoError(t, err)

		// Verify task is inactive
		var isActive bool
		err = db.QueryRow("SELECT is_active FROM task WHERE name = ?", task.Name).Scan(&isActive)
		require.NoError(t, err)
		assert.False(t, isActive)
	})

	t.Run("returns error when task not found", func(t *testing.T) {
		err := repo.Cancel(ctx, "non-existent-task")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found or already cancelled")
	})

	t.Run("returns error when task already cancelled", func(t *testing.T) {
		task := &core.StoredTask{
			ID:             uuid.MustParse("55555555-5555-5555-5555-555555555555"),
			Name:           "already-cancelled-task",
			OwnerSessionID: "session-123",
			Prompt:         "Test prompt",
			TriggerType:    "once",
			TriggerSpec:    "2025-01-01T00:00:00Z",
			IsActive:       false,
			CreatedAt:      time.Now(),
		}

		err := repo.Create(ctx, task)
		require.NoError(t, err)

		err = repo.Cancel(ctx, task.Name)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found or already cancelled")
	})
}

func TestScheduledTaskRepo_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTaskRepo(db)
	ctx := context.Background()

	t.Run("returns only active tasks", func(t *testing.T) {
		// Create active task
		activeTask := &core.StoredTask{
			ID:             uuid.MustParse("66666666-6666-6666-6666-666666666666"),
			Name:           "active-task",
			OwnerSessionID: "session-123",
			Prompt:         "Active prompt",
			TriggerType:    "interval",
			TriggerSpec:    "1h",
			IsActive:       true,
			CreatedAt:      time.Now().Add(-time.Hour),
		}
		err := repo.Create(ctx, activeTask)
		require.NoError(t, err)

		// Create inactive task
		inactiveTask := &core.StoredTask{
			ID:             uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			Name:           "inactive-task",
			OwnerSessionID: "session-123",
			Prompt:         "Inactive prompt",
			TriggerType:    "cron",
			TriggerSpec:    "0 0 * * *",
			IsActive:       false,
			CreatedAt:      time.Now(),
		}
		err = repo.Create(ctx, inactiveTask)
		require.NoError(t, err)

		tasks, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, "active-task", tasks[0].Name)
		assert.True(t, tasks[0].IsActive)
	})

	t.Run("returns tasks ordered by created_at DESC", func(t *testing.T) {
		// Clear existing data
		_, err := db.Exec("DELETE FROM task")
		require.NoError(t, err)

		// Create older task
		olderTask := &core.StoredTask{
			ID:             uuid.MustParse("88888888-8888-8888-8888-888888888888"),
			Name:           "older-task",
			OwnerSessionID: "session-123",
			Prompt:         "Older prompt",
			TriggerType:    "interval",
			TriggerSpec:    "1h",
			IsActive:       true,
			CreatedAt:      time.Now().Add(-2 * time.Hour),
		}
		err = repo.Create(ctx, olderTask)
		require.NoError(t, err)

		// Create newer task
		newerTask := &core.StoredTask{
			ID:             uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			Name:           "newer-task",
			OwnerSessionID: "session-123",
			Prompt:         "Newer prompt",
			TriggerType:    "once",
			TriggerSpec:    "2025-01-01T00:00:00Z",
			IsActive:       true,
			CreatedAt:      time.Now(),
		}
		err = repo.Create(ctx, newerTask)
		require.NoError(t, err)

		tasks, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, 2)
		assert.Equal(t, "newer-task", tasks[0].Name)
		assert.Equal(t, "older-task", tasks[1].Name)
	})

	t.Run("returns empty slice when no active tasks", func(t *testing.T) {
		// Clear existing data
		_, err := db.Exec("DELETE FROM task")
		require.NoError(t, err)

		tasks, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("scans null times correctly", func(t *testing.T) {
		task := &core.StoredTask{
			ID:             uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			Name:           "null-times-task",
			OwnerSessionID: "session-123",
			Prompt:         "Test prompt",
			TriggerType:    "cron",
			TriggerSpec:    "0 0 * * *",
			IsActive:       true,
			CreatedAt:      time.Now(),
			// LastRun and UpdatedAt are zero values (null in DB)
		}
		err := repo.Create(ctx, task)
		require.NoError(t, err)

		tasks, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.True(t, tasks[0].LastRun.IsZero())
		assert.Nil(t, tasks[0].UpdatedAt)
	})
}
