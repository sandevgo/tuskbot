package core

import (
	"context"
)

type Config interface {
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
