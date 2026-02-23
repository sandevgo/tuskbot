package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ModelCommand struct {
	ai core.AIProvider
}

func NewModelCommand(ai core.AIProvider) *ModelCommand {
	return &ModelCommand{
		ai: ai,
	}
}

func (c *ModelCommand) Name() string {
	return "model"
}

func (c *ModelCommand) Description() string {
	return "Show list of commands"
}

func (c *ModelCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	if len(args) == 0 {
		return fmt.Sprintf("Current model: %s\n", c.ai.GetModel()), nil
	}

	if err := c.ai.SetModel(args[0]); err != nil {
		return "", fmt.Errorf("failed to set model: %w", err)
	}

	return fmt.Sprintf("Model set to: %s\n", args[0]), nil
}
