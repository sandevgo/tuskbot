package config

import (
	"os"
	"path/filepath"
)

func GetRuntimePath() string {
	path := os.Getenv("TUSK_RUNTIME_PATH")
	if path == "" {
		path = ".tuskbot"
	}

	if !filepath.IsAbs(path) {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path)
	}
	return path
}
