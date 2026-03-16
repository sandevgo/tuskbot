package command

import (
	"github.com/sandevgo/tuskbot/internal/core"
)

func NewCommands(
	cfg core.ProviderConfig,
	state core.GlobalState,
	mcp core.MCPServer,
	swarm core.Swarm,
	memory core.Memory,
) []core.Command {
	// Create base commands
	cmds := []core.Command{
		NewModelCommand(cfg, state),
		NewMCPCommand(mcp),
		NewTasksCommand(swarm),
		NewStatsCommand(memory),
	}

	// Append HelpCommand (which needs the list of all commands)
	cmds = append(cmds, NewHelpCommand(cmds))

	return cmds
}
