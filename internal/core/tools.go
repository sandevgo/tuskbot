package core

import (
	"context"
	"encoding/json"
)

type NativeHandler func(context.Context, json.RawMessage) (string, error)

type NativeTool interface {
	GetDefinitions() map[string]struct {
		Description string
		Schema      string
		Handler     NativeHandler
	}
}

type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}
