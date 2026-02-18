package mcp

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/client"
	mcpproto "github.com/mark3labs/mcp-go/mcp"
	"github.com/sandevgo/tuskbot/internal/core"
)

type ConnectionPool interface {
	Add(ctx context.Context, name string, cfg ServerConfig) (*ManagedClient, error)
	Remove(name string) error
	Get(name string) (*ManagedClient, bool)
	All() map[string]*ManagedClient
	Close() error
}

type StdioConnectionPool struct {
	mu      sync.RWMutex
	clients map[string]*ManagedClient
}

func NewStdioConnectionPool() *StdioConnectionPool {
	return &StdioConnectionPool{
		clients: make(map[string]*ManagedClient),
	}
}

func (p *StdioConnectionPool) Add(ctx context.Context, name string, cfg ServerConfig) (*ManagedClient, error) {
	var env []string
	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// 2. Connect (Heavy I/O - done before locking)
	// The caller is responsible for providing a context with a timeout
	cli, err := client.NewStdioMCPClient(cfg.Command, env, cfg.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if err := cli.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	// 3. Initialize Protocol
	initReq := mcpproto.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcpproto.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcpproto.Implementation{
		Name:    core.TuskName,
		Version: core.TaskVersion,
	}
	initReq.Params.Capabilities = mcpproto.ClientCapabilities{}

	if _, err := cli.Initialize(ctx, initReq); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	// 4. Wrap and Store
	managed := &ManagedClient{
		Client: cli,
		name:   name,
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// If a client with this name already exists, close it first
	if old, exists := p.clients[name]; exists {
		// Close in background to avoid blocking the Add operation if the old process hangs
		go old.Close()
	}

	p.clients[name] = managed
	return managed, nil
}

func (p *StdioConnectionPool) Remove(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cli, exists := p.clients[name]; exists {
		delete(p.clients, name)
		return cli.Close()
	}
	return nil
}

func (p *StdioConnectionPool) Get(name string) (*ManagedClient, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	cli, ok := p.clients[name]
	return cli, ok
}

func (p *StdioConnectionPool) All() map[string]*ManagedClient {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to be safe
	result := make(map[string]*ManagedClient, len(p.clients))
	for k, v := range p.clients {
		result[k] = v
	}
	return result
}

func (p *StdioConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for _, cli := range p.clients {
		if err := cli.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	// Clear map
	p.clients = make(map[string]*ManagedClient)

	return errors.Join(errs...)
}
