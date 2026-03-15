package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sandevgo/tuskbot/configs"
)

type Migration func(runtimePath string) error

var Migrations = []Migration{
	setupInitialFiles,
}

func setupInitialFiles(p string) error {
	for _, dir := range []string{"config", "models", "prompt"} {
		if err := os.MkdirAll(filepath.Join(p, dir), 0755); err != nil {
			return err
		}
	}

	files := []string{"IDENTITY.md", "MEMORY.md", "SYSTEM.md", "USER.md", "mcp_config.json"}
	for _, name := range files {
		// Determine destination
		dest := filepath.Join(p, "prompt", name)
		if name == "mcp_config.json" {
			dest = filepath.Join(p, name)
		}

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

func Run(db *sql.DB, runtimePath string) error {
	var current int
	err := db.QueryRow("PRAGMA user_version").Scan(&current)
	if err != nil {
		return fmt.Errorf("failed to get user_version: %w", err)
	}

	for i := current; i < len(Migrations); i++ {
		if err := Migrations[i](runtimePath); err != nil {
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
