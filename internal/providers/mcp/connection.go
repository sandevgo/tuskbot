package mcp

import "context"

type ConnectionPool interface {
	Add(ctx context.Context, name string, cfg ServerConfig) (*ManagedClient, error)
	Remove(name string) error
	Get(name string) (*ManagedClient, bool)
	All() map[string]*ManagedClient
	Close() error
}
