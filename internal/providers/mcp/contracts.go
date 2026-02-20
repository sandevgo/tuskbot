package mcp

import "context"

type ConnectionPool interface {
	Add(ctx context.Context, name string, cfg ServerConfig) (*ManagedClient, error)
	Get(name string) (*ManagedClient, bool)
	Del(name string) error
	All() map[string]*ManagedClient
	Close() error
}
