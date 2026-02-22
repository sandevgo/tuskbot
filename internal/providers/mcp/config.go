package mcp

import "fmt"

type TransportType string

const (
	TransportHTTP  TransportType = "http"
	TransportStdio TransportType = "stdio"
)

type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig represents an entry in mcp_config.json
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (c *ServerConfig) GetTransport() (TransportType, error) {
	if c.URL != "" {
		return TransportHTTP, nil
	}
	if c.Command != "" {
		return TransportStdio, nil
	}
	return "", fmt.Errorf("invalid config: neither url nor command provided")
}
