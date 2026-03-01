package mcp

import (
	"encoding/json"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/mcp/tools"
)

func RegisterNativeTools(runtimePath string, swarm core.Swarm) (map[string]core.NativeHandler, []core.Tool) {
	handlers := make(map[string]core.NativeHandler)
	var defs []core.Tool

	register := func(t core.NativeTool) {
		for name, def := range t.GetDefinitions() {
			handlers[name] = def.Handler
			defs = append(defs, core.Tool{
				Type: "function",
				Function: core.Function{
					Name:        name,
					Description: def.Description,
					Parameters:  json.RawMessage(def.Schema),
				},
			})
		}
	}

	// Register Core Tools
	register(tools.NewFilesystem(runtimePath))
	register(tools.NewShell(runtimePath))
	register(tools.NewFetch())
	register(tools.NewSchedule(swarm))

	return handlers, defs
}
