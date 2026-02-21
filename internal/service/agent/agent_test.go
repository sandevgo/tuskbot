package agent

import (
	"context"
	"reflect"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

func TestSanitizeToolCalls(t *testing.T) {
	tests := []struct {
		name     string
		input    []core.Message
		expected []core.Message
	}{
		{
			name:     "empty messages",
			input:    []core.Message{},
			expected: nil,
		},
		{
			name: "normal conversation",
			input: []core.Message{
				{Role: core.RoleUser, Content: "hi"},
				{Role: core.RoleAssistant, Content: "calling tool", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result"},
			},
			expected: []core.Message{
				{Role: core.RoleUser, Content: "hi"},
				{Role: core.RoleAssistant, Content: "calling tool", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result"},
			},
		},
		{
			name: "orphaned tool call at start",
			input: []core.Message{
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result"},
				{Role: core.RoleUser, Content: "hi"},
			},
			expected: []core.Message{
				{Role: core.RoleUser, Content: "hi"},
			},
		},
		{
			name: "orphaned tool call after user message",
			input: []core.Message{
				{Role: core.RoleUser, Content: "hi"},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result"},
			},
			expected: []core.Message{
				{Role: core.RoleUser, Content: "hi"},
			},
		},
		{
			name: "tool call id mismatch",
			input: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tool", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleTool, ToolCallID: "call_2", Content: "result"},
			},
			expected: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tool", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
			},
		},
		{
			name: "multiple valid tool calls",
			input: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tools", ToolCalls: []core.ToolCall{{ID: "call_1"}, {ID: "call_2"}}},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result 1"},
				{Role: core.RoleTool, ToolCallID: "call_2", Content: "result 2"},
			},
			expected: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tools", ToolCalls: []core.ToolCall{{ID: "call_1"}, {ID: "call_2"}}},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result 1"},
				{Role: core.RoleTool, ToolCallID: "call_2", Content: "result 2"},
			},
		},
		{
			name: "mixed valid and invalid tool calls",
			input: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tools", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result 1"},
				{Role: core.RoleTool, ToolCallID: "call_2", Content: "result 2"}, // Invalid
			},
			expected: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tools", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result 1"},
			},
		},
		{
			name: "user message resets context",
			input: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tool", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleUser, Content: "interrupt"},
				{Role: core.RoleTool, ToolCallID: "call_1", Content: "result"}, // Now invalid because user interrupted
			},
			expected: []core.Message{
				{Role: core.RoleAssistant, Content: "calling tool", ToolCalls: []core.ToolCall{{ID: "call_1"}}},
				{Role: core.RoleUser, Content: "interrupt"},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeToolCalls(ctx, tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("sanitizeToolCalls() = %v, want %v", got, tt.expected)
			}
		})
	}
}
