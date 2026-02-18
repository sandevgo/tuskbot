package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const (
	sqlInsertMessage    = `INSERT INTO messages (session_id, role, content, tool_calls, tool_call_id) VALUES (?, ?, ?, ?, ?)`
	sqlSelectMessages   = `SELECT role, content, tool_calls, tool_call_id FROM messages WHERE session_id = ? ORDER BY id DESC LIMIT ?`
	sqlSelectUnembedded = `SELECT id, role, content, tool_calls, tool_call_id FROM messages WHERE embedded = false AND content != '' ORDER BY id ASC LIMIT ?`
	sqlInsertVector     = `INSERT INTO messages_vec (rowid, embedding) VALUES (?, ?)`
	sqlDeleteVector     = `DELETE FROM messages_vec WHERE rowid = ?`
	sqlMarkEmbedded     = `UPDATE messages SET embedded = true WHERE id = ?`
)

type MessagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) *MessagesRepo {
	return &MessagesRepo{db: db}
}

// AddMessage persists a message
func (r *MessagesRepo) AddMessage(ctx context.Context, sessionID string, msg core.Message) error {
	toolCallsStr, err := marshalToolCalls(msg.ToolCalls)
	if err != nil {
		return fmt.Errorf("failed to marshal tool calls: %w", err)
	}

	return r.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, sqlInsertMessage, sessionID, msg.Role, msg.Content, toolCallsStr, msg.ToolCallID)
		if err != nil {
			return fmt.Errorf("failed to insert message: %w", err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id: %w", err)
		}

		if len(msg.Embedding) > 0 {
			if err := r.persistEmbedding(tx, id, msg.Embedding[0]); err != nil {
				return fmt.Errorf("failed to persist embedding: %w", err)
			}
		}

		return nil
	})
}

// GetMessages retrieves the last 'limit' messages for a session in chronological order.
func (r *MessagesRepo) GetMessages(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
	rows, err := r.db.QueryContext(ctx, sqlSelectMessages, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	messages, err := scanMessages(rows)
	if err != nil {
		return nil, err
	}

	// Reverse to chronological order (oldest first)
	slices.Reverse(messages)

	log.FromCtx(ctx).Debug().Int("count", len(messages)).Msg("loaded history messages")
	return messages, nil
}

// GetUnembeddedMessages retrieves messages that haven't been embedded yet.
func (r *MessagesRepo) GetUnembeddedMessages(ctx context.Context, limit int) ([]core.StoredMessage, error) {
	rows, err := r.db.QueryContext(ctx, sqlSelectUnembedded, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	return scanStoredMessages(rows)
}

// UpdateMessageEmbedding updates the embedding for a specific message.
func (r *MessagesRepo) UpdateMessageEmbedding(ctx context.Context, id int64, embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("empty embedding provided")
	}

	return r.withTx(ctx, func(tx *sql.Tx) error {
		return r.persistEmbedding(tx, id, embedding)
	})
}

// withTx executes the given function within a transaction.
func (r *MessagesRepo) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// persistEmbedding saves the vector and marks the message as embedded.
func (r *MessagesRepo) persistEmbedding(tx *sql.Tx, msgID int64, embedding []float32) error {
	vecBlob, err := serializeVector(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize vector: %w", err)
	}

	if _, err := tx.Exec(sqlDeleteVector, msgID); err != nil {
		return fmt.Errorf("failed to delete vector: %w", err)
	}

	if _, err := tx.Exec(sqlInsertVector, msgID, vecBlob); err != nil {
		return fmt.Errorf("failed to insert vector: %w", err)
	}

	if _, err := tx.Exec(sqlMarkEmbedded, msgID); err != nil {
		return fmt.Errorf("failed to mark as embedded: %w", err)
	}

	return nil
}

// marshalToolCalls converts tool calls to JSON string, handling empty cases.
func marshalToolCalls(calls []core.ToolCall) (string, error) {
	if len(calls) == 0 {
		return "", nil
	}

	b, err := json.Marshal(calls)
	if err != nil {
		return "", err
	}

	// Handle "null" case for empty slices
	if string(b) == "null" {
		return "", nil
	}

	return string(b), nil
}

// unmarshalToolCalls parses JSON string into tool calls.
func unmarshalToolCalls(data string) ([]core.ToolCall, error) {
	if data == "" || data == "null" {
		return nil, nil
	}

	var calls []core.ToolCall
	if err := json.Unmarshal([]byte(data), &calls); err != nil {
		return nil, err
	}

	return calls, nil
}

// scanMessages scans rows into core.Message slices.
func scanMessages(rows *sql.Rows) ([]core.Message, error) {
	var messages []core.Message

	for rows.Next() {
		var msg core.Message
		var content, toolCallsStr, toolCallID sql.NullString

		if err := rows.Scan(&msg.Role, &content, &toolCallsStr, &toolCallID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Content = content.String
		msg.ToolCallID = toolCallID.String

		toolCalls, err := unmarshalToolCalls(toolCallsStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
		}
		msg.ToolCalls = toolCalls

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// scanStoredMessages scans rows into core.StoredMessage slices.
func scanStoredMessages(rows *sql.Rows) ([]core.StoredMessage, error) {
	var messages []core.StoredMessage

	for rows.Next() {
		var msg core.StoredMessage
		var content, toolCallsStr, toolCallID sql.NullString

		if err := rows.Scan(&msg.ID, &msg.Role, &content, &toolCallsStr, &toolCallID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Content = content.String
		msg.ToolCallID = toolCallID.String
		msg.ToolCalls = toolCallsStr.String

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
