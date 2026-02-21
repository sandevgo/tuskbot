package mcp

import (
	"fmt"
	"sync"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

func makeTools(names ...string) []core.Tool {
	tools := make([]core.Tool, len(names))
	for i, name := range names {
		tools[i] = core.Tool{Type: "function", Function: core.Function{Name: name}}
	}
	return tools
}

func makeRouting(pairs ...string) map[string]string {
	routing := make(map[string]string)
	for i := 0; i < len(pairs); i += 2 {
		routing[pairs[i]] = pairs[i+1]
	}
	return routing
}

func TestToolCache_Get(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(c *ToolCache)
		wantTools  int
		wantRoutes int
		wantOk     bool
	}{
		{
			name:       "empty_cache",
			setup:      func(c *ToolCache) {},
			wantTools:  0,
			wantRoutes: 0,
			wantOk:     false,
		},
		{
			name: "single_tool",
			setup: func(c *ToolCache) {
				c.Update(makeTools("tool1"), makeRouting("tool1", "server1"))
			},
			wantTools:  1,
			wantRoutes: 1,
			wantOk:     true,
		},
		{
			name: "multiple_tools",
			setup: func(c *ToolCache) {
				c.Update(makeTools("t1", "t2", "t3"), makeRouting("t1", "s1", "t2", "s2", "t3", "s3"))
			},
			wantTools:  3,
			wantRoutes: 3,
			wantOk:     true,
		},
		{
			name: "empty_update_marks_valid",
			setup: func(c *ToolCache) {
				c.Update([]core.Tool{}, map[string]string{})
			},
			wantTools:  0,
			wantRoutes: 0,
			wantOk:     true,
		},
		{
			name: "after_invalidate",
			setup: func(c *ToolCache) {
				c.Update(makeTools("tool"), makeRouting("tool", "server"))
				c.Invalidate()
			},
			wantTools:  0,
			wantRoutes: 0,
			wantOk:     false,
		},
		{
			name: "revalidate_after_invalidate",
			setup: func(c *ToolCache) {
				c.Update(makeTools("old"), makeRouting("old", "s1"))
				c.Invalidate()
				c.Update(makeTools("new"), makeRouting("new", "s2"))
			},
			wantTools:  1,
			wantRoutes: 1,
			wantOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()
			tt.setup(c)

			tools, routing, ok := c.Get()

			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
			if len(tools) != tt.wantTools {
				t.Errorf("tools count = %d, want %d", len(tools), tt.wantTools)
			}
			if len(routing) != tt.wantRoutes {
				t.Errorf("routing count = %d, want %d", len(routing), tt.wantRoutes)
			}
		})
	}
}

func TestToolCache_Update(t *testing.T) {
	tests := []struct {
		name    string
		updates []struct {
			tools   []core.Tool
			routing map[string]string
		}
		wantTools   []string
		wantRouting map[string]string
	}{
		{
			name: "single_update",
			updates: []struct {
				tools   []core.Tool
				routing map[string]string
			}{
				{makeTools("tool1"), makeRouting("tool1", "server1")},
			},
			wantTools:   []string{"tool1"},
			wantRouting: map[string]string{"tool1": "server1"},
		},
		{
			name: "overwrites_previous",
			updates: []struct {
				tools   []core.Tool
				routing map[string]string
			}{
				{makeTools("old1", "old2"), makeRouting("old1", "s1", "old2", "s2")},
				{makeTools("new"), makeRouting("new", "s3")},
			},
			wantTools:   []string{"new"},
			wantRouting: map[string]string{"new": "s3"},
		},
		{
			name: "nil_tools_becomes_empty",
			updates: []struct {
				tools   []core.Tool
				routing map[string]string
			}{
				{nil, makeRouting("key", "value")},
			},
			wantTools:   []string{},
			wantRouting: map[string]string{"key": "value"},
		},
		{
			name: "nil_routing_becomes_empty",
			updates: []struct {
				tools   []core.Tool
				routing map[string]string
			}{
				{makeTools("tool"), nil},
			},
			wantTools:   []string{"tool"},
			wantRouting: map[string]string{},
		},
		{
			name: "both_nil",
			updates: []struct {
				tools   []core.Tool
				routing map[string]string
			}{
				{nil, nil},
			},
			wantTools:   []string{},
			wantRouting: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()

			for _, u := range tt.updates {
				c.Update(u.tools, u.routing)
			}

			tools, routing, ok := c.Get()

			if !ok {
				t.Fatal("cache should be valid after update")
			}
			if len(tools) != len(tt.wantTools) {
				t.Errorf("tools count = %d, want %d", len(tools), len(tt.wantTools))
			}
			for i, name := range tt.wantTools {
				if tools[i].Function.Name != name {
					t.Errorf("tool[%d] = %s, want %s", i, tools[i].Function.Name, name)
				}
			}
			for k, v := range tt.wantRouting {
				if routing[k] != v {
					t.Errorf("routing[%s] = %s, want %s", k, routing[k], v)
				}
			}
			if len(routing) != len(tt.wantRouting) {
				t.Errorf("routing count = %d, want %d", len(routing), len(tt.wantRouting))
			}
		})
	}
}

func TestToolCache_Invalidate(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(c *ToolCache)
		wantOk bool
	}{
		{
			name:   "invalidate_empty_cache",
			setup:  func(c *ToolCache) { c.Invalidate() },
			wantOk: false,
		},
		{
			name: "invalidate_populated_cache",
			setup: func(c *ToolCache) {
				c.Update(makeTools("tool"), makeRouting("tool", "server"))
				c.Invalidate()
			},
			wantOk: false,
		},
		{
			name: "double_invalidate",
			setup: func(c *ToolCache) {
				c.Update(makeTools("tool"), makeRouting("tool", "server"))
				c.Invalidate()
				c.Invalidate()
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()
			tt.setup(c)

			_, _, ok := c.Get()
			if ok != tt.wantOk {
				t.Errorf("ok = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}

func TestToolCache_DeepCopy(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(tools []core.Tool, routing map[string]string)
		checkGet func(t *testing.T, tools []core.Tool, routing map[string]string)
	}{
		{
			name: "mutate_returned_tools",
			mutate: func(tools []core.Tool, routing map[string]string) {
				tools[0].Function.Name = "mutated"
			},
			checkGet: func(t *testing.T, tools []core.Tool, routing map[string]string) {
				if tools[0].Function.Name != "original" {
					t.Errorf("tool name = %s, want original", tools[0].Function.Name)
				}
			},
		},
		{
			name: "mutate_returned_routing",
			mutate: func(tools []core.Tool, routing map[string]string) {
				routing["hacked"] = "evil"
				delete(routing, "original")
			},
			checkGet: func(t *testing.T, tools []core.Tool, routing map[string]string) {
				if _, exists := routing["hacked"]; exists {
					t.Error("hacked key should not exist")
				}
				if routing["original"] != "server" {
					t.Errorf("original routing = %s, want server", routing["original"])
				}
			},
		},
		{
			name: "append_to_returned_tools",
			mutate: func(tools []core.Tool, routing map[string]string) {
				_ = append(tools, core.Tool{Function: core.Function{Name: "appended"}})
			},
			checkGet: func(t *testing.T, tools []core.Tool, routing map[string]string) {
				if len(tools) != 1 {
					t.Errorf("tools count = %d, want 1", len(tools))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()
			c.Update(makeTools("original"), makeRouting("original", "server"))

			tools, routing, _ := c.Get()
			tt.mutate(tools, routing)

			tools2, routing2, _ := c.Get()
			tt.checkGet(t, tools2, routing2)
		})
	}
}

func TestToolCache_UpdateDeepCopy(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(tools []core.Tool, routing map[string]string)
		checkGet func(t *testing.T, tools []core.Tool, routing map[string]string)
	}{
		{
			name: "mutate_source_tools",
			mutate: func(tools []core.Tool, routing map[string]string) {
				tools[0].Function.Name = "mutated"
			},
			checkGet: func(t *testing.T, tools []core.Tool, routing map[string]string) {
				if tools[0].Function.Name != "original" {
					t.Errorf("tool name = %s, want original", tools[0].Function.Name)
				}
			},
		},
		{
			name: "mutate_source_routing",
			mutate: func(tools []core.Tool, routing map[string]string) {
				routing["hacked"] = "evil"
				delete(routing, "original")
			},
			checkGet: func(t *testing.T, tools []core.Tool, routing map[string]string) {
				if _, exists := routing["hacked"]; exists {
					t.Error("hacked key should not exist")
				}
				if routing["original"] != "server" {
					t.Errorf("original routing = %s, want server", routing["original"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()

			tools := makeTools("original")
			routing := makeRouting("original", "server")
			c.Update(tools, routing)

			tt.mutate(tools, routing)

			cachedTools, cachedRouting, _ := c.Get()
			tt.checkGet(t, cachedTools, cachedRouting)
		})
	}
}

func TestToolCache_ConcurrentAccess(t *testing.T) {
	tests := []struct {
		name         string
		readers      int
		writers      int
		invalidators int
		iterations   int
	}{
		{
			name:         "light_load",
			readers:      5,
			writers:      2,
			invalidators: 1,
			iterations:   50,
		},
		{
			name:         "heavy_reads",
			readers:      20,
			writers:      2,
			invalidators: 1,
			iterations:   100,
		},
		{
			name:         "heavy_writes",
			readers:      5,
			writers:      10,
			invalidators: 2,
			iterations:   100,
		},
		{
			name:         "balanced",
			readers:      10,
			writers:      10,
			invalidators: 5,
			iterations:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()
			var wg sync.WaitGroup

			for i := 0; i < tt.writers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						c.Update(makeTools("tool"), makeRouting("tool", "server"))
					}
				}()
			}

			for i := 0; i < tt.readers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						c.Get()
					}
				}()
			}

			for i := 0; i < tt.invalidators; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						c.Invalidate()
					}
				}()
			}

			wg.Wait()
		})
	}
}

func TestToolCache_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		setup func(c *ToolCache)
		check func(t *testing.T, c *ToolCache)
	}{
		{
			name: "tools_with_same_name_different_servers",
			setup: func(c *ToolCache) {
				tools := makeTools("shared.tool", "shared.tool")
				routing := makeRouting("shared.tool", "server2") // last write wins
				c.Update(tools, routing)
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, ok := c.Get()
				if !ok {
					t.Fatal("cache should be valid")
				}
				if len(tools) != 2 {
					t.Errorf("tools count = %d, want 2", len(tools))
				}
				// Routing map can only hold one entry per key
				if len(routing) != 1 {
					t.Errorf("routing count = %d, want 1", len(routing))
				}
			},
		},
		{
			name: "empty_tool_name",
			setup: func(c *ToolCache) {
				c.Update(makeTools(""), makeRouting("", "server"))
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, ok := c.Get()
				if !ok {
					t.Fatal("cache should be valid")
				}
				if tools[0].Function.Name != "" {
					t.Error("empty name should be preserved")
				}
				if routing[""] != "server" {
					t.Error("empty key routing should work")
				}
			},
		},
		{
			name: "routing_without_matching_tool",
			setup: func(c *ToolCache) {
				tools := makeTools("tool1")
				routing := makeRouting("tool1", "s1", "orphan", "s2")
				c.Update(tools, routing)
			},
			check: func(t *testing.T, c *ToolCache) {
				_, routing, _ := c.Get()
				if routing["orphan"] != "s2" {
					t.Error("orphan routing should be preserved")
				}
			},
		},
		{
			name: "tool_without_routing",
			setup: func(c *ToolCache) {
				tools := makeTools("tool1", "tool2")
				routing := makeRouting("tool1", "server1")
				c.Update(tools, routing)
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, _ := c.Get()
				if len(tools) != 2 {
					t.Errorf("tools count = %d, want 2", len(tools))
				}
				if routing["tool2"] != "" {
					t.Errorf("tool2 routing = %s, want empty", routing["tool2"])
				}
			},
		},
		{
			name: "large_dataset",
			setup: func(c *ToolCache) {
				names := make([]string, 1000)
				pairs := make([]string, 2000)
				for i := 0; i < 1000; i++ {
					names[i] = fmt.Sprintf("tool%d", i)
					pairs[i*2] = names[i]
					pairs[i*2+1] = fmt.Sprintf("server%d", i%10)
				}
				c.Update(makeTools(names...), makeRouting(pairs...))
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, ok := c.Get()
				if !ok {
					t.Fatal("cache should be valid")
				}
				if len(tools) != 1000 {
					t.Errorf("tools count = %d, want 1000", len(tools))
				}
				if len(routing) != 1000 {
					t.Errorf("routing count = %d, want 1000", len(routing))
				}
			},
		},
		{
			name: "special_characters_in_names",
			setup: func(c *ToolCache) {
				c.Update(
					makeTools("tool/with/slashes", "tool.with.dots", "tool:with:colons", "tool with spaces"),
					makeRouting(
						"tool/with/slashes", "s1",
						"tool.with.dots", "s2",
						"tool:with:colons", "s3",
						"tool with spaces", "s4",
					),
				)
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, ok := c.Get()
				if !ok {
					t.Fatal("cache should be valid")
				}
				if len(tools) != 4 {
					t.Errorf("tools count = %d, want 4", len(tools))
				}
				if routing["tool/with/slashes"] != "s1" {
					t.Error("slash routing failed")
				}
				if routing["tool with spaces"] != "s4" {
					t.Error("space routing failed")
				}
			},
		},
		{
			name: "unicode_names",
			setup: func(c *ToolCache) {
				c.Update(
					makeTools("工具", "инструмент", "🔧"),
					makeRouting("工具", "服务器", "инструмент", "сервер", "🔧", "🖥️"),
				)
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, ok := c.Get()
				if !ok {
					t.Fatal("cache should be valid")
				}
				if len(tools) != 3 {
					t.Errorf("tools count = %d, want 3", len(tools))
				}
				if routing["工具"] != "服务器" {
					t.Error("chinese routing failed")
				}
				if routing["🔧"] != "🖥️" {
					t.Error("emoji routing failed")
				}
			},
		},
		{
			name:  "get_on_fresh_cache_multiple_times",
			setup: func(c *ToolCache) {},
			check: func(t *testing.T, c *ToolCache) {
				for i := 0; i < 3; i++ {
					_, _, ok := c.Get()
					if ok {
						t.Errorf("iteration %d: fresh cache should return ok=false", i)
					}
				}
			},
		},
		{
			name: "rapid_invalidate_update_cycle",
			setup: func(c *ToolCache) {
				for i := 0; i < 100; i++ {
					c.Update(makeTools("tool"), makeRouting("tool", "server"))
					c.Invalidate()
				}
				c.Update(makeTools("final"), makeRouting("final", "last"))
			},
			check: func(t *testing.T, c *ToolCache) {
				tools, routing, ok := c.Get()
				if !ok {
					t.Fatal("cache should be valid")
				}
				if len(tools) != 1 || tools[0].Function.Name != "final" {
					t.Error("expected final tool")
				}
				if routing["final"] != "last" {
					t.Error("expected final routing")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewToolCache()
			tt.setup(c)
			tt.check(t, c)
		})
	}
}
