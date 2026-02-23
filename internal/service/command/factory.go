package command

import (
	"github.com/sandevgo/tuskbot/internal/core"
)

func NewCommands(
	ai core.AIProvider,
	mcp core.MCPServer,
) []core.Command {
	return []core.Command{
		NewModelCommand(ai),
		NewMCPCommand(mcp),
	}
}
