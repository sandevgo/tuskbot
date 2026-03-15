package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sandevgo/tuskbot/configs"
	"github.com/sandevgo/tuskbot/internal/core"
)

type Migration func(cfg config) error

var Migrations = []Migration{
	checkEmbeddingModel,
	setupInitialFiles,
}

type config interface {
	core.AppConfig
	core.PromptConfig
	core.EmbeddingConfig
}

func checkEmbeddingModel(cfg config) error {
	if _, err := os.Stat(cfg.GetEmbeddingModelPath()); os.IsNotExist(err) {
		return fmt.Errorf(
			"embedding model not found at %s\nRun 'tusk install' to download it",
			cfg.GetEmbeddingModelPath(),
		)
	}
	return nil
}

func setupInitialFiles(cfg config) error {
	dirs := []string{
		filepath.Dir(cfg.GetConfigPath()),
		filepath.Dir(cfg.GetPromptPath()),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	fileMap := map[string]string{
		"IDENTITY.md":     cfg.GetIdentityPath(),
		"MEMORY.md":       cfg.GetMemoryPath(),
		"SYSTEM.md":       cfg.GetSystemPath(),
		"USER.md":         cfg.GetUserProfilePath(),
		"mcp_config.json": cfg.GetMCPConfigPath(),
	}

	for name, dest := range fileMap {
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			content, err := configs.FS.ReadFile(name)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dest, content, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func Run(db *sql.DB, cfg config) error {
	var current int
	err := db.QueryRow("PRAGMA user_version").Scan(&current)
	if err != nil {
		return fmt.Errorf("failed to get user_version: %w", err)
	}

	for i := current; i < len(Migrations); i++ {
		if err := Migrations[i](cfg); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}

		newVersion := i + 1
		_, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", newVersion))
		if err != nil {
			return fmt.Errorf("failed to set user_version to %d: %w", newVersion, err)
		}
	}
	return nil
}
