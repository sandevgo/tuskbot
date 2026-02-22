package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type ConnectionPool interface {
	Add(ctx context.Context, name string, cfg ServerConfig) (*ManagedClient, error)
	Del(name string) error
	Get(name string) (*ManagedClient, bool)
	All() map[string]*ManagedClient
	Close() error
}

var _ ConnectionPool = (*Pool)(nil)

type TransportFactory func(TransportType) (Transport, error)

type Pool struct {
	mu               sync.RWMutex
	clients          map[string]*ManagedClient
	transportFactory TransportFactory
}

func NewPool() *Pool {
	return &Pool{
		clients:          make(map[string]*ManagedClient),
		transportFactory: NewTransport,
	}
}

func NewPoolWithFactory(factory TransportFactory) *Pool {
	return &Pool{
		clients:          make(map[string]*ManagedClient),
		transportFactory: factory,
	}
}

func (p *Pool) Add(ctx context.Context, name string, cfg ServerConfig) (*ManagedClient, error) {
	tType, err := cfg.GetTransport()
	if err != nil {
		return nil, err
	}

	transport, err := p.transportFactory(tType)
	if err != nil {
		return nil, err
	}

	client, err := transport(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("transport creation failed: %w", err)
	}

	managed := &ManagedClient{
		Client: client,
		name:   name,
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if old, exists := p.clients[name]; exists {
		go old.Close()
	}

	p.clients[name] = managed
	return managed, nil
}

func (p *Pool) Del(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cli, exists := p.clients[name]; exists {
		delete(p.clients, name)
		return cli.Close()
	}
	return nil
}

func (p *Pool) Get(name string) (*ManagedClient, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	cli, ok := p.clients[name]
	return cli, ok
}

func (p *Pool) All() map[string]*ManagedClient {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]*ManagedClient, len(p.clients))
	for k, v := range p.clients {
		result[k] = v
	}
	return result
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for _, cli := range p.clients {
		if err := cli.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	p.clients = make(map[string]*ManagedClient)

	return errors.Join(errs...)
}
