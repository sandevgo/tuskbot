package core

import (
	"context"
)

type AppConfig interface {
	GetRuntimePath() string
	GetSystemPath() string
	GetIdentityPath() string
	GetUserProfilePath() string
	GetMemoryPath() string
	GetDatabasePath() string
	GetMCPConfigPath() string
	GetModel() string
	GetProvider() string
	GetAnthropicAPIKey() string
	GetOpenAIAPIKey() string
	GetOpenRouterAPIKey() string
	GetOllamaAPIKey() string
	GetOllamaBaseURL() string
	GetCustomOpenAIBaseURL() string
	GetCustomOpenAIAPIKey() string
	IsTelegramSelected() bool
}

type GlobalState interface {
	ChangeModel(ctx context.Context, model string) error
}
