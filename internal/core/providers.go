package core

import "context"

type AIProvider interface {
	Chat(ctx context.Context, req ChatRequest) (Message, error)
	Models(ctx context.Context) ([]Model, error)
	Capabilities() ProviderCapabilities
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
	Messages []Message
	Tools    []Tool
	Cache    CachePolicy
}

func NewChatRequest(promptCtx PromptContext, tools []Tool) ChatRequest {
	req := ChatRequest{
		Messages: promptCtx.Messages,
		Tools:    tools,
	}

	if promptCtx.StaticPrefixCount <= 0 {
		return req
	}

	breakpoints := make([]int, promptCtx.StaticPrefixCount)
	for i := range promptCtx.StaticPrefixCount {
		breakpoints[i] = i
	}

	req.Cache.Prompt = &PromptCachePolicy{
		Mode:               PromptCacheModePrefer,
		MessageBreakpoints: breakpoints,
		IncludeTools:       len(tools) > 0,
	}

	return req
}

type CachePolicy struct {
	Prompt     *PromptCachePolicy
	ModelReuse *ModelReusePolicy
}

type PromptCacheMode string

const (
	PromptCacheModeDefault PromptCacheMode = "default"
	PromptCacheModePrefer  PromptCacheMode = "prefer"
	PromptCacheModeBypass  PromptCacheMode = "bypass"
)

type PromptCachePolicy struct {
	Mode               PromptCacheMode
	MessageBreakpoints []int
	IncludeTools       bool
}

type ModelReuseMode string

const (
	ModelReuseModeDefault ModelReuseMode = "default"
	ModelReuseModePrefer  ModelReuseMode = "prefer"
	ModelReuseModeBypass  ModelReuseMode = "bypass"
)

type ModelReusePolicy struct {
	Mode ModelReuseMode
	TTL  string
}

type PromptCacheSupport string

const (
	PromptCacheSupportNone      PromptCacheSupport = "none"
	PromptCacheSupportAutomatic PromptCacheSupport = "automatic"
	PromptCacheSupportExplicit  PromptCacheSupport = "explicit"
)

type ProviderCapabilities struct {
	PromptCache PromptCacheSupport
	ModelReuse  bool
}
