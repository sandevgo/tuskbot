package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type SystemServiceConfig struct {
	Name         string `env:"TUSK_SERVICE_NAME" envDefault:"tuskbot"`
	DisplayName  string `env:"TUSK_SERVICE_DISPLAY_NAME" envDefault:"TuskBot"`
	Description  string `env:"TUSK_SERVICE_DESCRIPTION" envDefault:"TuskBot background agent service"`
	UserService  bool   `env:"TUSK_SERVICE_USER_MODE" envDefault:"true"`
	SandboxMode  bool   `env:"TUSK_SERVICE_SANDBOX_MODE" envDefault:"false"`
	LogDirectory string `env:"TUSK_SERVICE_LOG_DIRECTORY"`
	Path         string `env:"TUSK_SERVICE_PATH"`
}

func NewSystemServiceConfig(ctx context.Context) *SystemServiceConfig {
	c := &SystemServiceConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse system service config")
	}
	return c
}

func (c *SystemServiceConfig) GetLogDirectory() string {
	if c.LogDirectory != "" {
		return c.LogDirectory
	}
	return filepath.Join(GetRuntimePath(), "logs")
}

func (c *SystemServiceConfig) GetPATH() string {
	if strings.TrimSpace(c.Path) != "" {
		return c.Path
	}

	home, _ := os.UserHomeDir()
	pathParts := []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, ".npm-global", "bin"),
		filepath.Join(home, ".bun", "bin"),
		filepath.Join(home, ".nvm", "current", "bin"),
		filepath.Join(home, ".fnm", "current", "bin"),
		filepath.Join(home, ".local", "share", "pnpm"),
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
	}

	if currentPath := strings.TrimSpace(os.Getenv("PATH")); currentPath != "" {
		pathParts = append(pathParts, currentPath)
	}

	return strings.Join(pathParts, ":")
}
