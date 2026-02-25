package command

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ModelCommand struct {
	cfg       core.ProviderConfig
	state     core.GlobalState
	formatter *ResponseFormatter
}

func NewModelCommand(
	cfg core.ProviderConfig,
	state core.GlobalState,
) *ModelCommand {
	return &ModelCommand{
		cfg:       cfg,
		state:     state,
		formatter: NewResponseFormatter(),
	}
}

func (c *ModelCommand) Name() string {
	return "model"
}

func (c *ModelCommand) Description() string {
	return "Show or change current model"
}

func (c *ModelCommand) Execute(ctx context.Context, sessionID string, args []string) (string, error) {
	if len(args) == 0 {
		return c.formatter.Combine(
			c.formatter.Info("Current Model"),
			c.formatter.Label("Provider", c.cfg.GetProvider()),
			c.formatter.Label("Model", c.cfg.GetModel()),
			c.formatter.Usage("/model [provider]/[model]"),
			c.formatter.Examples([]string{
				"/model openai/gpt-4",
				"/model anthropic/claude-3-sonnet",
				"/model openrouter/openai/gpt-3.5-turbo",
			}),
		), nil
	}

	if err := c.state.ChangeModel(ctx, args[0]); err != nil {
		return "", fmt.Errorf("failed to set model: %w", err)
	}

	return c.formatter.Combine(
		c.formatter.Success(fmt.Sprintf("Model changed to: `%s/%s`", c.cfg.GetProvider(), c.cfg.GetModel())),
	), nil
}
