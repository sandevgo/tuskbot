package env

import (
	"os"
	"path/filepath"
	"strings"
)

const maxServicePathLength = 4096

func BuildServicePATH(customPath string) string {
	pathParts := make([]string, 0, 24)

	if customPath = strings.TrimSpace(customPath); customPath != "" {
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

		normalized := filepath.Clean(p)
		if _, ok := seen[normalized]; ok {
			continue
		}

		nextLength := currentLength + len(p)
		if len(result) > 0 {
			nextLength += len(pathSep)
		}
		if nextLength > maxServicePathLength {
			continue
		}

		seen[normalized] = struct{}{}
		result = append(result, p)
		currentLength = nextLength
	}

	return strings.Join(result, pathSep)
}
