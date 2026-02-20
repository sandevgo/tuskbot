package tools

import (
	"context"
	"encoding/json"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/mcp"
)

type tool interface {
	GetDefinitions() map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}
}

func RegisterNativeTools(
	runtimePath string,
	registry *mcp.Registry,
	pool mcp.ConnectionPool,
	cache *mcp.ToolCache,
) (map[string]mcp.NativeHandler, []core.Tool) {
	handlers := make(map[string]mcp.NativeHandler)
	var defs []core.Tool

	register := func(t tool) {
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
	register(NewManageTool(registry, pool, cache))
	register(NewFilesystem(runtimePath))
	register(NewShell(runtimePath))
	register(NewFetch())

	return handlers, defs
}
