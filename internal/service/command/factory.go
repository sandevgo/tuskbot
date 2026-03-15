package command

import (
	"github.com/sandevgo/tuskbot/internal/core"
)

func NewCommands(
	cfg core.ProviderConfig,
	state core.GlobalState,
	mcp core.MCPServer,
	swarm core.Swarm,
) []core.Command {
	return []core.Command{
		NewModelCommand(cfg, state),
		NewMCPCommand(mcp),
		NewTasksCommand(swarm),
	}
}
