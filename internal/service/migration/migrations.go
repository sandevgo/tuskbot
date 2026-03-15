package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sandevgo/tuskbot/configs"
)

const versionFile = ".version"

// Migration defines a function that performs a filesystem or config change
type Migration func(runtimePath string) error

var Migrations = []Migration{
	// Version 0: Initial Setup
	func(p string) error {
		// 1. Create directories
		for _, dir := range []string{"prompt", "data", "models"} {
			if err := os.MkdirAll(filepath.Join(p, dir), 0755); err != nil {
				return err
			}
		}

		// 2. Copy embedded files if they don't exist
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
	},
}

func Run(runtimePath string) error {
	current := getCurrentVersion(runtimePath)
	for i := current; i < len(Migrations); i++ {
		if err := Migrations[i](runtimePath); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
		if err := saveVersion(runtimePath, i+1); err != nil {
			return err
		}
	}
	return nil
}

func getCurrentVersion(runtimePath string) int {
	data, err := os.ReadFile(filepath.Join(runtimePath, versionFile))
	if err != nil {
		return 0
	}
	v, _ := strconv.Atoi(string(data))
	return v
}

func saveVersion(runtimePath string, version int) error {
	return os.WriteFile(filepath.Join(runtimePath, versionFile), []byte(strconv.Itoa(version)), 0644)
}
