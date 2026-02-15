package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type AppConfig struct {
	MainModel         string `env:"TUSK_MAIN_MODEL,required,notEmpty"`
	ChatChannel       string `env:"TUSK_CHAT_CHANNEL,required,notEmpty"`
	ContextWindowSize int    `env:"TUSK_CONTEXT_WINDOW_SIZE" envDefault:"30"`

	AnthropicAPIKey  string `env:"TUSK_ANTHROPIC_API_KEY"`
	OpenAIAPIKey     string `env:"TUSK_OPENAI_API_KEY"`
	OpenRouterAPIKey string `env:"TUSK_OPENROUTER_API_KEY"`
	OllamaAPIKey     string `env:"TUSK_OLLAMA_API_KEY"`
	OllamaBaseURL    string `env:"TUSK_OLLAMA_BASE_URL" envDefault:"http://127.0.0.1:11434"`

	runtimePath string
	Provider    string
	Model       string
}

func NewAppConfig(ctx context.Context) *AppConfig {
	c := &AppConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse App config")
	}

	i := strings.Index(c.MainModel, "/")
	if i > 0 {
		c.Provider = c.MainModel[:i]
		c.Model = c.MainModel[i+1:]
	}

	c.runtimePath = GetRuntimePath()
	return c
}

func (c *AppConfig) GetRuntimePath() string {
	return c.runtimePath
}

func (c *AppConfig) GetSystemPath() string {
	return filepath.Join(c.runtimePath, "SYSTEM.md")
}

func (c *AppConfig) GetIdentityPath() string {
	return filepath.Join(c.runtimePath, "IDENTITY.md")
}

func (c *AppConfig) GetUserProfilePath() string {
	return filepath.Join(c.runtimePath, "USER.md")
}

func (c *AppConfig) GetMemoryPath() string {
	return filepath.Join(c.runtimePath, "MEMORY.md")
}

func (c *AppConfig) GetDatabasePath() string {
	return filepath.Join(c.runtimePath, "tuskbot.db")
}

func (c *AppConfig) GetMCPConfigPath() string {
	return filepath.Join(c.runtimePath, "mcp_config.json")
}

func (c *AppConfig) IsTelegramSelected() bool {
	return strings.ToLower(c.ChatChannel) == "telegram"
}

// GetRuntimePath can be used outside AppConfig
func GetRuntimePath() string {
	path := os.Getenv("TUSK_RUNTIME_PATH")
	if path == "" {
		path = ".tuskbot"
	}

	if !filepath.IsAbs(path) {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path)
	}
	return path
}
