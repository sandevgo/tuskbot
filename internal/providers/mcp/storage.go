package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sandevgo/tuskbot/pkg/log"
)

type FileStorage struct {
	path string
	mu   sync.RWMutex
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

// Load reads the config. If the file is missing, it creates a default one.
func (c *FileStorage) Load(ctx context.Context) (*Config, error) {
	return c.load(ctx, true)
}

// load reads the config. If createIfMissing is false, it returns an error for missing files.
func (c *FileStorage) load(ctx context.Context, createIfMissing bool) (*Config, error) {
	c.mu.RLock()
	data, err := os.ReadFile(c.path)
	c.mu.RUnlock()

	if err != nil {
		if os.IsNotExist(err) {
			if !createIfMissing {
				return nil, fmt.Errorf("config file not found: %w", err)
			}

			dir := filepath.Dir(c.path)
			if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
				return nil, fmt.Errorf("config directory does not exist: %w", err)
			}

			log.FromCtx(ctx).Info().Msg("mcp_config.json not found, creating default")

			config := &Config{
				MCPServers: make(map[string]ServerConfig),
			}

			if err = c.Save(ctx, config); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return config, nil
		}
		return nil, fmt.Errorf("failed to read mcp config: %w", err)
	}

	config := &Config{
		MCPServers: make(map[string]ServerConfig),
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse mcp config: %w", err)
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]ServerConfig)
	}

	return config, nil
}

func (c *FileStorage) Save(ctx context.Context, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

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

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Read file directly - this is atomic
				c.mu.RLock()
				data, err := os.ReadFile(c.path)
				c.mu.RUnlock()

				if err != nil {
					lastMod = time.Time{}
					continue
				}

				info, err = os.Stat(c.path)
				if err != nil {
					lastMod = time.Time{}
					continue
				}

				if !info.ModTime().After(lastMod) {
					continue
				}

				var config Config
				if err := json.Unmarshal(data, &config); err != nil {
					log.FromCtx(ctx).Error().Err(err).Msg("failed to parse mcp config")
					continue
				}

				if config.MCPServers == nil {
					config.MCPServers = make(map[string]ServerConfig)
				}

				lastMod = info.ModTime()

				select {
				case updates <- config:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return updates, nil
}
