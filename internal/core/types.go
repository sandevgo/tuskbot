package core

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
	RoleTool      = "tool"
)

const (
	TuskName          = "TuskBot"
	TuskUserAgent     = "TuskBot-Agent/0.1"
	TuskRepositoryURL = "https://github.com/sandevgo/tuskbot"
	TaskVersion       = "0.1.0"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Reasoning  string     `json:"reasoning,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`

	// Internal
	Embedding [][]float32 `json:"-"`
}
