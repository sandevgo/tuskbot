package config

import (
	"context"

	"github.com/caarlos0/env/v9"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type SystemServiceConfig struct {
	Name        string `env:"TUSK_SERVICE_NAME" envDefault:"tuskbot"`
	DisplayName string `env:"TUSK_SERVICE_DISPLAY_NAME" envDefault:"TuskBot"`
	Description string `env:"TUSK_SERVICE_DESCRIPTION" envDefault:"TuskBot background agent service"`
	UserService bool   `env:"TUSK_SERVICE_USER_MODE" envDefault:"true"`
	SandboxMode bool   `env:"TUSK_SERVICE_SANDBOX_MODE" envDefault:"true"`
}

func NewSystemServiceConfig(ctx context.Context) *SystemServiceConfig {
	c := &SystemServiceConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse system service config")
	}
	return c
}
