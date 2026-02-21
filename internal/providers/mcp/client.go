package mcp

import (
	"sync"

	"github.com/mark3labs/mcp-go/client"
)

type ManagedClient struct {
	*client.Client
	mu     sync.RWMutex
	closed bool
	name   string
}

func (mc *ManagedClient) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	if mc.closed {
		return nil
	}
	mc.closed = true
	if mc.Client == nil {
		return nil
	}
	return mc.Client.Close()
}

func (mc *ManagedClient) IsClosed() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.closed
}
