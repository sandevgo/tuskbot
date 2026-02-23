package core

import (
	"context"
)

type AppConfig interface {
	GetRuntimePath() string
	GetDatabasePath() string
	GetMCPConfigPath() string
	GetContextWindowSize() int
	IsTelegramSelected() bool
}

type PromptConfig interface {
	GetSystemPath() string
	GetIdentityPath() string
	GetUserProfilePath() string
	GetMemoryPath() string
}

type ProviderConfig interface {
	GetModel() string
	SetModel(model string) error
	GetProvider() string
	GetAnthropicAPIKey() string
	GetOpenAIAPIKey() string
	GetOpenRouterAPIKey() string
	GetOllamaAPIKey() string
	GetOllamaBaseURL() string
	GetCustomOpenAIBaseURL() string
	GetCustomOpenAIAPIKey() string
}

type EmbeddingConfig interface {
	GetEmbeddingModel() string
}

type TelegramConfig interface {
	GetTelegramToken() string
	GetTelegramOwnerID() int64
}

type GlobalState interface {
	ChangeModel(ctx context.Context, model string) error
}
