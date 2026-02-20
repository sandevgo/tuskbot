package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sandevgo/tuskbot/pkg/log"
)

// ServerConfig represents an entry in mcp_config.json
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

type Storage interface {
	Load(ctx context.Context) (*Config, error)
	Save(ctx context.Context, cfg *Config) error
	Watch(ctx context.Context) (<-chan Config, error)
}

type FileStorage struct {
	path string
	mu   sync.RWMutex
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

func (c *FileStorage) Load(ctx context.Context) (*Config, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	config := &Config{
		MCPServers: make(map[string]ServerConfig),
	}

	data, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			log.FromCtx(ctx).Info().Msg("mcp_config.json not found, creating default")

			// Save empty (default) config
			if err = c.Save(ctx, config); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read mcp config: %w", err)
		}
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mcp config: %w", err)
	}

	return config, nil
}

func (c *FileStorage) Save(ctx context.Context, cfg *Config) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (c *FileStorage) Watch(ctx context.Context) (<-chan Config, error) {
	updates := make(chan Config)

	info, err := os.Stat(c.path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}
	lastMod := info.ModTime()

	go func() {
		defer close(updates)

		// Poll for changes every second
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				info, err = os.Stat(c.path)
				if err != nil {
					// File might be temporarily missing during atomic write or deleted
					continue
				}

				if info.ModTime().After(lastMod) {
					lastMod = info.ModTime()
					newCfg, err := c.Load(ctx)
					if err != nil {
						log.FromCtx(ctx).Error().Err(err).Msg("failed to reload mcp config")
						continue
					}

					select {
					case updates <- *newCfg:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return updates, nil
}
