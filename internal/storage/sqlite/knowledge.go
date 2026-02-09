package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type KnowledgeRepo struct {
	db *sql.DB
}

func NewKnowledgeRepo(db *sql.DB) *KnowledgeRepo {
	return &KnowledgeRepo{db: db}
}

func (r *KnowledgeRepo) SaveFact(ctx context.Context, fact core.StoredKnowledge) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	vecBlob, err := serializeVector(fact.Embedding)
	if err != nil {
		return err
	}

	// 1. Insert Metadata
	res, err := tx.ExecContext(ctx,
		`INSERT INTO knowledge (fact, category, source, fact_hash) VALUES (?, ?, ?, ?)`,
		fact.Fact, fact.Category, fact.Source, fact.FactHash,
	)
	if err != nil {
		return fmt.Errorf("failed to insert knowledge metadata: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// 2. Insert Vector into Virtual Table using rowid
	_, err = tx.ExecContext(ctx,
		`INSERT INTO knowledge_vec (rowid, embedding) VALUES (?, ?)`,
		id, vecBlob,
	)
	if err != nil {
		return fmt.Errorf("failed to insert knowledge vector: %w", err)
	}

	return tx.Commit()
}

func (r *KnowledgeRepo) SearchContext(ctx context.Context, vector []float32, limitKnowledge, limitHistory int) ([]core.ContextItem, error) {
	vecBlob, err := serializeVector(vector)
	if err != nil {
		return nil, err
	}

	var results []core.ContextItem

	// 1. Search Knowledge
	// Note: vec_distance_L2 is standard for sqlite-vec. Lower is better.
	kQuery := `
		SELECT
			k.id, k.fact, k.source, k.created_at, v.distance
		FROM knowledge_vec v
		JOIN knowledge k ON k.id = v.rowid
		WHERE v.embedding MATCH ? AND k = ?
		ORDER BY v.distance
	`
	rows, err := r.db.QueryContext(ctx, kQuery, vecBlob, limitKnowledge)
	if err != nil {
		return nil, fmt.Errorf("knowledge search failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item core.ContextItem
		item.Type = "fact"
		if err := rows.Scan(&item.ID, &item.Content, &item.Source, &item.CreatedAt, &item.Score); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	// 2. Search Messages
	mQuery := `
		SELECT
			m.id, m.content, m.role, m.created_at, v.distance
		FROM messages_vec v
		JOIN messages m ON m.id = v.rowid
		WHERE v.embedding MATCH ? AND k = ?
		ORDER BY v.distance
	`
	mRows, err := r.db.QueryContext(ctx, mQuery, vecBlob, limitHistory)
	if err != nil {
		return nil, fmt.Errorf("message search failed: %w", err)
	}
	defer mRows.Close()

	for mRows.Next() {
		var item core.ContextItem
		var role string
		item.Type = "message"
		item.Source = "history"
		if err := mRows.Scan(&item.ID, &item.Content, &role, &item.CreatedAt, &item.Score); err != nil {
			return nil, err
		}
		// Format content to include role for context clarity
		item.Content = fmt.Sprintf("%s: %s", strings.ToUpper(role), item.Content)
		results = append(results, item)
	}

	return results, nil
}

func (r *KnowledgeRepo) MarkMessagesExtracted(ctx context.Context, messageIDs []int64) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Simple query builder for IN clause
	query := fmt.Sprintf("UPDATE messages SET extracted = 1 WHERE id IN (%s)",
		strings.Trim(strings.Replace(fmt.Sprint(messageIDs), " ", ",", -1), "[]"))

	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *KnowledgeRepo) GetUnextractedMessages(ctx context.Context, limit int) ([]core.StoredMessage, error) {
	query := `
		SELECT id, session_id, role, content, created_at 
		FROM messages 
		WHERE extracted = 0 AND role != 'system' AND role != 'tool'
		ORDER BY id ASC 
		LIMIT ?`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []core.StoredMessage
	for rows.Next() {
		var m core.StoredMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}
