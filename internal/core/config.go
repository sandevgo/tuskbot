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
	IsTelegramSelected()
}

type GlobalState interface {
	ChangeModel(ctx context.Context, model string) error
}
