package mcp

import (
	"context"
	"encoding/json"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/mcp/tools"
)

type tool interface {
	GetDefinitions() map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}
}

func RegisterNativeTools(runtimePath string) (map[string]NativeHandler, []core.Tool) {
	handlers := make(map[string]NativeHandler)
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
	register(tools.NewFilesystem(runtimePath))
	register(tools.NewShell(runtimePath))
	register(tools.NewFetch())

	return handlers, defs
}
