package mcp

import (
	"context"
	"sync"
)

type Storage interface {
	Load(ctx context.Context) (*Config, error)
	Save(ctx context.Context, cfg *Config) error
	Watch(ctx context.Context) (<-chan Config, error)
}

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

	// Create new state
	newServers := make(map[string]ServerConfig, len(r.servers)+1)
	for k, v := range r.servers {
		newServers[k] = v
	}
	newServers[name] = cfg

	// Try to save first
	if err := r.storage.Save(ctx, &Config{MCPServers: newServers}); err != nil {
		return err
	}

	// Only update in-memory state if save succeeded
	r.servers = newServers
	return nil
}

func (r *Registry) Remove(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create new state without the removed server
	newServers := make(map[string]ServerConfig, len(r.servers))
	for k, v := range r.servers {
		if k != name {
			newServers[k] = v
		}
	}

	// Try to save first
	if err := r.storage.Save(ctx, &Config{MCPServers: newServers}); err != nil {
		return err
	}

	// Only update in-memory state if save succeeded
	r.servers = newServers
	return nil
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

func (r *Registry) Watch(ctx context.Context) (<-chan Config, error) {
	ch, err := r.storage.Watch(ctx)
	if err != nil {
		return nil, err
	}

	out := make(chan Config)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case cfg, ok := <-ch:
				if !ok {
					return
				}

				r.mu.Lock()
				if cfg.MCPServers == nil {
					r.servers = make(map[string]ServerConfig)
				} else {
					r.servers = cfg.MCPServers
				}
				r.mu.Unlock()

				select {
				case out <- cfg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}
