package mcp

import (
	"sync"

	"github.com/sandevgo/tuskbot/internal/core"
)

type ToolCache struct {
	mu           sync.RWMutex
	version      uint64
	tools        []core.Tool
	toolToServer map[string]string // tool name -> server name
	valid        bool
}

func NewToolCache() *ToolCache {
	return &ToolCache{
		toolToServer: make(map[string]string),
	}
}

func (c *ToolCache) Get() (tools []core.Tool, routing map[string]string, version uint64, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.valid {
		return nil, nil, c.version, false
	}

	// Deep copy to prevent external mutation
	toolsCopy := make([]core.Tool, len(c.tools))
	copy(toolsCopy, c.tools)

	routingCopy := make(map[string]string, len(c.toolToServer))
	for k, v := range c.toolToServer {
		routingCopy[k] = v
	}

	return toolsCopy, routingCopy, c.version, true
}

func (c *ToolCache) Update(tools []core.Tool, routing map[string]string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.version++
	c.valid = true

	// Store copies to prevent external mutation of our internal state
	c.tools = make([]core.Tool, len(tools))
	copy(c.tools, tools)

	c.toolToServer = make(map[string]string, len(routing))
	for k, v := range routing {
		c.toolToServer[k] = v
	}

	return c.version
}

// Invalidate marks cache stale and returns new version
func (c *ToolCache) Invalidate() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.version++
	c.valid = false
	c.tools = nil
	c.toolToServer = make(map[string]string)

	return c.version
}

func (c *ToolCache) IsValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.valid
}
