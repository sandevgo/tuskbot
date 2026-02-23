package core

import "context"

type Command interface {
	Name() string
	Description() string
	Execute(ctx context.Context, sessionID string, args []string) (string, error)
}
