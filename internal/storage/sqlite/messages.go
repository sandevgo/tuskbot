package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type MessagesRepo struct {
	db *sql.DB
}

func NewMessagesRepo(db *sql.DB) *MessagesRepo {
	return &MessagesRepo{db: db}
}

func (h *MessagesRepo) AddMessage(ctx context.Context, sessionID string, msg core.Message) error {
	toolCallsJSON, err := json.Marshal(msg.ToolCalls)
	if err != nil {
		return fmt.Errorf("failed to marshal tool calls: %w", err)
	}

	// If ToolCalls is "null" (empty slice), store as empty string to save space
	tcStr := string(toolCallsJSON)
	if tcStr == "null" {
		tcStr = ""
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert into main table
	query := `INSERT INTO messages (session_id, role, content, tool_calls, tool_call_id) VALUES (?, ?, ?, ?, ?)`
	res, err := tx.ExecContext(ctx, query, sessionID, msg.Role, msg.Content, tcStr, msg.ToolCallID)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}

	// 2. Insert into vector table if embedding exists
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if msg.Embedding != nil && len(msg.Embedding) > 0 {
		vecBlob, err := serializeVector(msg.Embedding[0])
		if err != nil {
			return err
		}
		// Use 'rowid' explicitly to ensure the vector is tied to the message ID
		_, err = tx.ExecContext(ctx, `INSERT INTO messages_vec (rowid, embedding) VALUES (?, ?)`, id, vecBlob)
		if err != nil {
			return fmt.Errorf("failed to insert message vector: %w", err)
		}
	}

	return tx.Commit()
}

func (h *MessagesRepo) GetMessages(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
	// Fetch the LAST 'limit' messages by ordering DESC
	query := `SELECT role, content, tool_calls, tool_call_id FROM messages WHERE session_id = ? ORDER BY id DESC LIMIT ?`

	rows, err := h.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []core.Message
	for rows.Next() {
		var msg core.Message
		var content, toolCallsStr, toolCallID sql.NullString

		// Use NullString to safely handle potential NULLs in DB
		if err := rows.Scan(&msg.Role, &content, &toolCallsStr, &toolCallID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Content = content.String
		msg.ToolCallID = toolCallID.String

		if toolCallsStr.Valid && toolCallsStr.String != "" && toolCallsStr.String != "null" {
			if err := json.Unmarshal([]byte(toolCallsStr.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// The query returned messages in Reverse Chronological Order (Newest -> Oldest).
	// We need to reverse them back to Chronological Order (Oldest -> Newest) for the LLM.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	log.FromCtx(ctx).Debug().Int("count", len(messages)).Msg("loaded history messages")
	return messages, nil
}

func (h *MessagesRepo) GetUnembeddedMessages(ctx context.Context, limit int) ([]core.StoredMessage, error) {
	query := `
		SELECT role, content, tool_calls, tool_call_id
		FROM messages 
		WHERE embedded = false AND content != '' ORDER BY id DESC LIMIT ?
	`

	rows, err := h.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []core.StoredMessage
	for rows.Next() {
		var msg core.StoredMessage
		var content, toolCallsStr, toolCallID sql.NullString

		if err := rows.Scan(&msg.Role, &content, &toolCallsStr, &toolCallID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Content = content.String
		msg.ToolCallID = toolCallID.String

		if toolCallsStr.Valid && toolCallsStr.String != "" && toolCallsStr.String != "null" {
			if err := json.Unmarshal([]byte(toolCallsStr.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (h *MessagesRepo) UpdateMessageEmbedding(ctx context.Context, id int64, embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("empty embedding provided")
	}

	vecBlob, err := serializeVector(embedding)
	if err != nil {
		return fmt.Errorf("serialize vector: %w", err)
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert or update vector (use INSERT...ON CONFLICT for atomic upsert)
	_, err = tx.ExecContext(ctx, `
        INSERT INTO messages_vec (rowid, embedding)
        VALUES (?, ?)
        ON CONFLICT(rowid) DO UPDATE SET embedding = excluded.embedding
    `, id, vecBlob)
	if err != nil {
		return fmt.Errorf("insert vector: %w", err)
	}

	// 2. Mark as embedded in main table (CRITICAL - was missing!)
	_, err = tx.ExecContext(ctx,
		`UPDATE messages SET embedded = true WHERE id = ?`,
		id)
	if err != nil {
		return fmt.Errorf("update embedded flag: %w", err)
	}

	return tx.Commit()
}
