package core

import "context"

type UpdateFunc func(Message)

type Agent interface {
	Run(ctx context.Context, sessionID string, input string, onUpdate UpdateFunc) (string, error)
	Notify(ctx context.Context, task *Task, result string) error
}

type SubAgent interface {
	Run(ctx context.Context, task *Task, onComplete UpdateFunc) (string, error)
}

type ToolExecutor interface {
	Execute(ctx context.Context, toolCalls []ToolCall) []Message
}
