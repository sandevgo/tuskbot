package core

import "context"

type CommandRouter interface {
	GetCommands() ([]Command, error)
}

type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, sessionID string, args []string) (string, error)
}
