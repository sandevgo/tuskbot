package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ModelCommand struct {
	appCfg core.AppConfig
	state  core.GlobalState
}

func NewModelCommand(
	appCfg core.AppConfig,
	state core.GlobalState,
) *ModelCommand {
	return &ModelCommand{
		appCfg: appCfg,
		state:  state,
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
		return fmt.Sprintf("Current model: %s\n", c.appCfg.GetModel()), nil
	}

	if err := c.state.ChangeModel(ctx, args[0]); err != nil {
		return "", fmt.Errorf("failed to set model: %w", err)
	}

	return fmt.Sprintf("Model set to: %s\n", c.appCfg.GetModel()), nil
}
