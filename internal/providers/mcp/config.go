package mcp

import "fmt"

type TransportType string

const (
	TransportHTTP  TransportType = "http"
	TransportSSE   TransportType = "sse"
	TransportStdio TransportType = "stdio"
)

type Config struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig represents an entry in mcp_config.json
type ServerConfig struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Type    TransportType     `json:"type,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (c *ServerConfig) GetTransport() (TransportType, error) {
	if c.URL != "" {
		switch c.Type {
		case TransportSSE:
			return TransportSSE, nil
		case TransportHTTP:
			return TransportHTTP, nil
		default:
			return "", fmt.Errorf("unknown transport type for URL: %s (use http or sse)", c.Type)
		}
	}
	if c.Command != "" {
		return TransportStdio, nil
	}
	return "", fmt.Errorf("invalid config: neither url nor command provided")
}
