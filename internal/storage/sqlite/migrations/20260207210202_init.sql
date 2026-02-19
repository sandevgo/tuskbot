-- +goose Up
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL DEFAULT 'default',
    role TEXT NOT NULL,
    content TEXT,
    tool_calls TEXT,
    tool_call_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    embedded BOOLEAN DEFAULT FALSE,
    extracted BOOLEAN DEFAULT 0
);

CREATE VIRTUAL TABLE messages_vec USING vec0(
    embedding float[768]
);

CREATE TABLE knowledge (
    id INTEGER PRIMARY KEY,
    fact TEXT NOT NULL,
    category TEXT NOT NULL, -- 'preference', 'user_fact', 'project', 'instruction'
    source TEXT, -- 'extracted:conv_123' | 'manual' | 'file:project.md'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE VIRTUAL TABLE knowledge_vec USING vec0(
    embedding float[768] -- 768 for e5-base
);

-- +goose Down
DROP TABLE knowledge_vec;
DROP TABLE knowledge;
DROP TABLE messages_vec;
DROP TABLE messages;
