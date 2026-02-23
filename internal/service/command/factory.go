package command

import (
	"github.com/sandevgo/tuskbot/internal/core"
)

func NewCommands(
	appCfg core.AppConfig,
	state core.GlobalState,
	mcp core.MCPServer,
) []core.Command {
	return []core.Command{
		NewModelCommand(appCfg, state),
		NewMCPCommand(mcp),
	}
}
