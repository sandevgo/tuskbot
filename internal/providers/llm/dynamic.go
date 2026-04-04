package llm

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/sandevgo/tuskbot/internal/core"
)

type DynamicProvider struct {
	config  core.ProviderConfig
	current atomic.Value
	mu      sync.RWMutex
}

func NewDynamicProvider(
	ctx context.Context,
	config core.ProviderConfig,
) (*DynamicProvider, error) {
	d := &DynamicProvider{
		config: config,
	}

	provider, err := NewProvider(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial provider: %w", err)
	}

	d.current.Store(provider)
	return d, nil
}

func (d *DynamicProvider) Chat(ctx context.Context, req core.ChatRequest) (core.Message, error) {
	provider := d.current.Load().(core.AIProvider)
	return provider.Chat(ctx, req)
}

func (d *DynamicProvider) Models(ctx context.Context) ([]core.Model, error) {
	provider := d.current.Load().(core.AIProvider)
	return provider.Models(ctx)
}

func (d *DynamicProvider) Capabilities() core.ProviderCapabilities {
	provider := d.current.Load().(core.AIProvider)
	return provider.Capabilities()
}

// GetModel (thread-safe)
func (d *DynamicProvider) GetModel() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config.GetModel()
}

func (d *DynamicProvider) SetModel(ctx context.Context, model string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Update config (persist)
	if err := d.config.SetModel(model); err != nil {
		return err
	}

	// Create new provider
	newProvider, err := NewProvider(ctx, d.config)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Atomic swap
	d.current.Store(newProvider)
	return nil
}
