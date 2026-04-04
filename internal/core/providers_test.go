package core

import (
	"reflect"
	"testing"
)

func TestNewChatRequest(t *testing.T) {
	tests := []struct {
		name            string
		ctx             PromptContext
		tools           []Tool
		wantBreakpoints []int
	}{
		{
			name: "no static prefix",
			ctx: PromptContext{
				Messages: []Message{{Role: RoleUser, Content: "hi"}},
			},
		},
		{
			name: "static prefix without tools",
			ctx: PromptContext{
				Messages:          []Message{{Role: RoleSystem, Content: "sys"}, {Role: RoleUser, Content: "hi"}},
				StaticPrefixCount: 1,
			},
			wantBreakpoints: []int{0},
		},
		{
			name: "static prefix with tools",
			ctx: PromptContext{
				Messages: []Message{
					{Role: RoleSystem, Content: "sys"},
					{Role: RoleSystem, Content: "identity"},
					{Role: RoleUser, Content: "hi"},
				},
				StaticPrefixCount: 2,
			},
			tools:           []Tool{{Type: "function"}},
			wantBreakpoints: []int{0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewChatRequest(tt.ctx, tt.tools)

			if !reflect.DeepEqual(got.Messages, tt.ctx.Messages) {
				t.Fatalf("Messages = %#v, want %#v", got.Messages, tt.ctx.Messages)
			}
			if !reflect.DeepEqual(got.Tools, tt.tools) {
				t.Fatalf("Tools = %#v, want %#v", got.Tools, tt.tools)
			}
			if !reflect.DeepEqual(got.CacheBreakpoints, tt.wantBreakpoints) {
				t.Fatalf("CacheBreakpoints = %#v, want %#v", got.CacheBreakpoints, tt.wantBreakpoints)
			}
		})
	}
}
