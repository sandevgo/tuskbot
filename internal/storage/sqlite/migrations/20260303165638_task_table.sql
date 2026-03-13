-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
