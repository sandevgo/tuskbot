package mcp

import "context"

type TransportType string

const (
	TransportHTTP  TransportType = "http"
	TransportStdio TransportType = "stdio"
)

type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

type Storage interface {
	Load(ctx context.Context) (*Config, error)
	Save(ctx context.Context, cfg *Config) error
	Watch(ctx context.Context) (<-chan Config, error)
}

// ServerConfig represents an entry in mcp_config.json
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	URL     string            `json:"serverUrl,omitempty"`
}

func (c *ServerConfig) GetTransport() TransportType {
	if c.URL != "" {
		return TransportHTTP
	}
	return TransportStdio
}
