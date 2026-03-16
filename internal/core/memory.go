package core

import (
	"context"
	"time"
)

type TokenCounter interface {
	CountTokens(text string) int
}

type Memory interface {
	GetFullContext(ctx context.Context, sessionID, userQuery string) ([]Message, error)
	GetTaskContext(ctx context.Context, sessionID, instruction string) ([]Message, error)
	SaveMessage(ctx context.Context, sessionID string, msg Message) error
	CountTokens(messages []Message) int
}

// ContextItem represents a piece of retrieved information (either a Fact or a past Message)
type ContextItem struct {
	ID        int64
	Content   string
	Type      string // "fact" or "message"
	Score     float32
	Source    string // "extracted" or "history"
	CreatedAt time.Time
}
