package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

const modelResponse = "Current model: %s/%s\n\nTo change use: /model <provider>/<model>"

type ModelCommand struct {
	cfg   core.ProviderConfig
	state core.GlobalState
}

func NewModelCommand(
	cfg core.ProviderConfig,
	state core.GlobalState,
) *ModelCommand {
	return &ModelCommand{
		cfg:   cfg,
		state: state,
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
		return fmt.Sprintf(modelResponse, c.cfg.GetProvider(), c.cfg.GetModel()), nil
	}

	if err := c.state.ChangeModel(ctx, args[0]); err != nil {
		return "", fmt.Errorf("failed to set model: %w", err)
	}

	return fmt.Sprintf("Model set to: %s\n", c.cfg.GetModel()), nil
}
