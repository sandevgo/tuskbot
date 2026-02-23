package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/config"
)

type ModelCommand struct {
	cfg *config.AppConfig
}

func NewModelCommand(cfg *config.AppConfig) *ModelCommand {
	return &ModelCommand{
		cfg: cfg,
	}
}

func (c *ModelCommand) Name() string {
	return "model"
}

func (c *ModelCommand) Description() string {
	return "Show list of commands"
}

func (c *ModelCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	return fmt.Sprintf("Current model: %s\n", c.cfg.MainModel), nil
}
