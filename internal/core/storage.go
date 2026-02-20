package core

import (
	"context"
	"time"
)

type MessagesRepository interface {
	AddMessage(ctx context.Context, sessionID string, msg Message) error
	GetMessages(ctx context.Context, sessionID string, limit int) ([]Message, error)
	GetUnembeddedMessages(ctx context.Context, limit int) ([]StoredMessage, error)
	UpdateMessageEmbedding(ctx context.Context, id int64, embedding []float32) error
}

type KnowledgeRepository interface {
	SaveFact(ctx context.Context, fact StoredKnowledge) error
	SearchContext(ctx context.Context, vector []float32, limitKnowledge, limitHistory int) ([]ContextItem, error)
	MarkMessagesExtracted(ctx context.Context, messageIDs []int64) error
	GetUnextractedMessages(ctx context.Context, limit int) ([]StoredMessage, error)
	GetRecentExtractedMessages(ctx context.Context, limit int, before time.Time, threshold time.Duration) ([]StoredMessage, error)
}

type StoredMessage struct {
	ID         int64     `json:"id"`
	SessionID  string    `json:"session_id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ToolCalls  string    `json:"tool_calls,omitempty"`
	ToolCallID string    `json:"tool_call_id,omitempty"`
	Embedding  []float32 `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	Embedded   bool      `json:"embedded"`
	Extracted  bool      `json:"extracted"`
}

type StoredKnowledge struct {
	ID        int64      `json:"id"`
	Fact      string     `json:"fact"`
	Category  string     `json:"category"`
	Source    string     `json:"source"`
	Embedding []float32  `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
