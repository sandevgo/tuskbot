package core

import "context"

type AIProvider interface {
	Chat(ctx context.Context, req ChatRequest) (Message, error)
	Models(ctx context.Context) ([]Model, error)
}

type Embedder interface {
	EncodeQuery(ctx context.Context, text string) ([]float32, error)
	EncodePassage(ctx context.Context, text string) ([][]float32, error)
}

type EmbeddingModel interface {
	Dims() int
	GetURL() string
	GetModelName() string
	Shutdown() error
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

type ChatRequest struct {
	Messages         []Message
	Tools            []Tool
	CachePrefixCount int
}

func NewChatRequest(promptCtx PromptContext, tools []Tool) ChatRequest {
	req := ChatRequest{
		Messages: promptCtx.Messages,
		Tools:    tools,
	}

	if promptCtx.StaticPrefixCount <= 0 {
		return req
	}

	req.CachePrefixCount = promptCtx.StaticPrefixCount

	return req
}
