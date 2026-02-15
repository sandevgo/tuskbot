package core

import "context"

type AIProvider interface {
	Chat(ctx context.Context, history []Message, tools []Tool) (Message, error)
	Models(ctx context.Context) ([]Model, error)
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([][]float32, error)
}

type MCPServer interface {
	GetTools(ctx context.Context) ([]Tool, error)
	CallTool(ctx context.Context, name string, args string) (string, error)
}

type Model struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ContextLength int    `json:"context_length"`
}
