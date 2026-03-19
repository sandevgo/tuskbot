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

const maxServicePathLength = 4096

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
	pathParts := make([]string, 0, 24)

	if customPath := strings.TrimSpace(c.Path); customPath != "" {
		pathParts = append(pathParts, filepath.SplitList(customPath)...)
	} else {
		home := strings.TrimSpace(os.Getenv("HOME"))
		if userHome, err := os.UserHomeDir(); err == nil && strings.TrimSpace(userHome) != "" {
			home = strings.TrimSpace(userHome)
		}

		if home != "" {
			pathParts = append(pathParts,
				filepath.Join(home, ".local", "bin"),
				filepath.Join(home, ".npm-global", "bin"),
				filepath.Join(home, ".bun", "bin"),
				filepath.Join(home, ".nvm", "current", "bin"),
				filepath.Join(home, ".fnm", "current", "bin"),
				filepath.Join(home, ".local", "share", "pnpm"),
			)
		}

		pathParts = append(pathParts,
			"/usr/local/sbin",
			"/usr/local/bin",
			"/usr/sbin",
			"/usr/bin",
			"/sbin",
			"/bin",
		)

		if currentPath := strings.TrimSpace(os.Getenv("PATH")); currentPath != "" {
			pathParts = append(pathParts, filepath.SplitList(currentPath)...)
		}
	}

	pathSep := string(os.PathListSeparator)
	seen := make(map[string]struct{}, len(pathParts))
	result := make([]string, 0, len(pathParts))
	currentLength := 0

	for _, p := range pathParts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}

		nextLength := currentLength + len(p)
		if len(result) > 0 {
			nextLength += len(pathSep)
		}
		if nextLength > maxServicePathLength {
			continue
		}

		seen[p] = struct{}{}
		result = append(result, p)
		currentLength = nextLength
	}

	return strings.Join(result, pathSep)
}
