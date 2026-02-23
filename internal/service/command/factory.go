package command

import (
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
)

func NewCommands(
	cfg *config.AppConfig,
	mcp core.MCPServer,
) []core.Command {
	return []core.Command{
		NewModelCommand(cfg),
		NewMCPCommand(mcp),
	}
}
