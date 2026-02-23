package core

import "context"

type CmdRouter interface {
	Execute(ctx context.Context, sessionID, input string) (string, bool)
}

type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, sessionID string, args []string) (string, error)
}
