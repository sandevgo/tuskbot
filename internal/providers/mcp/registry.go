package mcp

import (
	"context"
	"sync"
)

type Registry struct {
	storage Storage
	mu      sync.RWMutex
	servers map[string]ServerConfig
}

func NewRegistry(storage Storage) *Registry {
	return &Registry{
		storage: storage,
		servers: make(map[string]ServerConfig),
	}
}

func (r *Registry) Load(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, err := r.storage.Load(ctx)
	if err != nil {
		return err
	}

	r.servers = cfg.MCPServers
	return nil
}

func (r *Registry) Add(ctx context.Context, name string, cfg ServerConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.servers[name] = cfg
	return r.save(ctx)
}

func (r *Registry) Remove(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.servers, name)
	return r.save(ctx)
}

func (r *Registry) Get(name string) (ServerConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, ok := r.servers[name]
	return cfg, ok
}

func (r *Registry) List() map[string]ServerConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return copy
	result := make(map[string]ServerConfig, len(r.servers))
	for k, v := range r.servers {
		result[k] = v
	}
	return result
}

func (r *Registry) save(ctx context.Context) error {
	return r.storage.Save(ctx, &Config{
		MCPServers: r.servers,
	})
}
