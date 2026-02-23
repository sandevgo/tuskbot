package config

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

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

	CustomOpenAIBaseURL string `env:"TUSK_CUSTOM_OPENAI_BASE_URL"`
	CustomOpenAIAPIKey  string `env:"TUSK_CUSTOM_OPENAI_API_KEY"`

	// runtime state
	mu          sync.Mutex
	runtimePath string
	provider    string
	model       string
}

func NewAppConfig(ctx context.Context, runtimePath string) *AppConfig {
	c := &AppConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse App config")
	}

	i := strings.Index(c.MainModel, "/")
	if i > 0 {
		c.provider = c.MainModel[:i]
		c.model = c.MainModel[i+1:]
	}

	c.runtimePath = runtimePath
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

func (c *AppConfig) GetContextWindowSize() int {
	return c.ContextWindowSize
}

func (c *AppConfig) IsTelegramSelected() bool {
	return strings.ToLower(c.ChatChannel) == "telegram"
}

func (c *AppConfig) GetAnthropicAPIKey() string {
	return c.AnthropicAPIKey
}

func (c *AppConfig) GetOpenAIAPIKey() string {
	return c.OpenAIAPIKey
}

func (c *AppConfig) GetOpenRouterAPIKey() string {
	return c.OpenRouterAPIKey
}

func (c *AppConfig) GetOllamaAPIKey() string {
	return c.OllamaAPIKey
}

func (c *AppConfig) GetOllamaBaseURL() string {
	return c.OllamaBaseURL
}

func (c *AppConfig) GetCustomOpenAIBaseURL() string {
	return c.CustomOpenAIBaseURL
}

func (c *AppConfig) GetCustomOpenAIAPIKey() string {
	return c.CustomOpenAIAPIKey
}

func (c *AppConfig) GetModel() string {
	return c.model
}

func (c *AppConfig) SetModel(model string) {
	c.model = model
}

func (c *AppConfig) GetProvider() string {
	return c.provider
}

func (c *AppConfig) SetProvider(provider string) {
	c.provider = provider
}

func (c *AppConfig) Persist(ctx context.Context) error {
	return fmt.Errorf("persistance not implemented")
}
