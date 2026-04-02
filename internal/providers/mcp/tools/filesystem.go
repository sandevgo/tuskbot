package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sandevgo/tuskbot/internal/core"
)

const readFileSchema = `
{
  "type": "object",
  "properties": {
    "path": { "type": "string", "description": "The path to the file to read" }
  },
  "required": ["path"]
}
`

const writeFileSchema = `
{
  "type": "object",
  "properties": {
    "path": { "type": "string", "description": "The path to the file to write" },
    "content": { "type": "string", "description": "The content to write to the file" }
  },
  "required": ["path", "content"]
}
`

const editFileSchema = `
{
  "type": "object",
  "properties": {
    "path": { "type": "string", "description": "The path to the file to edit" },
    "find": { "type": "string", "description": "The exact string to find in the file" },
    "replace": { "type": "string", "description": "The string to replace it with" }
  },
  "required": ["path", "find", "replace"]
}
`

const listDirSchema = `
{
  "type": "object",
  "properties": {
    "path": { "type": "string", "description": "The directory path to list" }
  },
  "required": ["path"]
}
`

const searchFilesSchema = `
{
  "type": "object",
  "properties": {
    "path": { "type": "string", "description": "The directory or file path to search in" },
    "query": { "type": "string", "description": "The string to search for" }
  },
  "required": ["path", "query"]
}
`

const getFileInfoSchema = `
{
  "type": "object",
  "properties": {
    "path": { "type": "string", "description": "The path to the file or directory to inspect" }
  },
  "required": ["path"]
}
`

const globSchema = `
{
  "type": "object",
  "properties": {
    "pattern": { "type": "string", "description": "The glob pattern to match files (e.g., *.go, **/*.ts)" },
    "path": { "type": "string", "description": "The directory to search in (defaults to current directory)" }
  },
  "required": ["pattern"]
}
`

type Filesystem struct {
	BasePath string
}

func NewFilesystem(basePath string) *Filesystem {
	if basePath == "" {
		basePath, _ = os.Getwd()
	}
	return &Filesystem{BasePath: basePath}
}

func (fs *Filesystem) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(fs.BasePath, p)
}

func (fs *Filesystem) ReadFile(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	path := fs.resolvePath(input.Path)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

func (fs *Filesystem) WriteFile(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	path := fs.resolvePath(input.Path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directories: %w", err)
	}

	if err := os.WriteFile(path, []byte(input.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return fmt.Sprintf("Successfully wrote to %s", input.Path), nil
}

func (fs *Filesystem) EditFile(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Path    string `json:"path"`
		Find    string `json:"find"`
		Replace string `json:"replace"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	path := fs.resolvePath(input.Path)
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	content := string(contentBytes)

	if !strings.Contains(content, input.Find) {
		return "", fmt.Errorf("exact string not found in file")
	}

	// Replace all occurrences
	newContent := strings.ReplaceAll(content, input.Find, input.Replace)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully edited %s", input.Path), nil
}

func (fs *Filesystem) ListDir(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	path := fs.resolvePath(input.Path)
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to list directory: %w", err)
	}

	var result string
	for _, entry := range entries {
		info, _ := entry.Info()
		prefix := "[FILE]"
		if entry.IsDir() {
			prefix = "[DIR] "
		}
		result += fmt.Sprintf("%s %s (%d bytes)\n", prefix, entry.Name(), info.Size())
	}
	return result, nil
}

func (fs *Filesystem) SearchFiles(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Path  string `json:"path"`
		Query string `json:"query"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	searchPath := fs.resolvePath(input.Path)
	var results strings.Builder
	matchCount := 0

	// Walk the directory
	err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			return nil // Skip errors accessing files
		}

		// Skip hidden directories and common vendor dirs
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." && d.Name() != ".." {
				return filepath.SkipDir
			}
			if d.Name() == "vendor" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Read file
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		// Check if binary (read first 512 bytes)
		// Simple heuristic: check for null byte
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		for i := 0; i < n; i++ {
			if buf[i] == 0 {
				return nil // Skip binary file
			}
		}

		// Reset file pointer
		file.Seek(0, 0)

		// Scan line by line
		scanner := bufio.NewScanner(file)

		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Check if valid UTF-8
			if !utf8.ValidString(line) {
				continue
			}

			if strings.Contains(line, input.Query) {
				// Limit output line length
				displayLine := strings.TrimSpace(line)
				if len(displayLine) > 200 {
					displayLine = displayLine[:200] + "..."
				}

				// Use relative path for display if possible
				displayPath := path
				if rel, err := filepath.Rel(fs.BasePath, path); err == nil {
					displayPath = rel
				}

				results.WriteString(fmt.Sprintf("%s:%d: %s\n", displayPath, lineNum, displayLine))
				matchCount++
				if matchCount >= 100 {
					results.WriteString("... (too many matches, stopping search)\n")
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	if matchCount == 0 {
		return "No matches found.", nil
	}

	return results.String(), nil
}

func (fs *Filesystem) GetFileInfo(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	path := fs.resolvePath(input.Path)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	return fmt.Sprintf(
		"Path: %s\nSize: %d bytes\nIsDir: %t\nMode: %s\nModTime: %s\n",
		input.Path,
		info.Size(),
		info.IsDir(),
		info.Mode(),
		info.ModTime().Format(time.RFC3339),
	), nil
}

// Glob finds files matching a pattern. It supports standard glob patterns
// and ** for recursive matching.
func (fs *Filesystem) Glob(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	searchPath := fs.BasePath
	if input.Path != "" {
		searchPath = fs.resolvePath(input.Path)
	}

	var matches []string

	// Check if pattern contains ** for recursive search
	if strings.Contains(input.Pattern, "**") {
		err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			// Get relative path for matching
			relPath, err := filepath.Rel(searchPath, path)
			if err != nil {
				return nil
			}

			if matchGlobRecursive(relPath, input.Pattern) {
				matches = append(matches, path)
			}
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("glob walk failed: %w", err)
		}
	} else {
		// Standard non-recursive glob
		pattern := filepath.Join(searchPath, input.Pattern)
		m, err := filepath.Glob(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid glob pattern: %w", err)
		}
		matches = m
	}

	if len(matches) == 0 {
		return "No files matched the pattern.", nil
	}

	var result strings.Builder
	for _, match := range matches {
		rel, err := filepath.Rel(fs.BasePath, match)
		if err != nil {
			rel = match
		}
		result.WriteString(rel + "\n")
	}

	return result.String(), nil
}

// matchGlobRecursive handles simple ** patterns
func matchGlobRecursive(path, pattern string) bool {
	// Normalize paths
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)

	// Split pattern by **
	parts := strings.Split(pattern, "**")

	// If no **, use standard match (handled by caller usually, but good fallback)
	if len(parts) == 1 {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	currentPath := path
	for i, part := range parts {
		if part == "" {
			continue
		}

		if i == 0 {
			// First part must be prefix
			if !strings.HasPrefix(currentPath, part) {
				return false
			}
			currentPath = strings.TrimPrefix(currentPath, part)
		} else if i == len(parts)-1 {
			// Last part must be suffix
			if !strings.HasSuffix(currentPath, part) {
				return false
			}
			currentPath = strings.TrimSuffix(currentPath, part)
		} else {
			// Middle part must be found
			idx := strings.Index(currentPath, part)
			if idx == -1 {
				return false
			}
			currentPath = currentPath[idx+len(part):]
		}
	}
	return true
}

func (fs *Filesystem) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     core.NativeHandler
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     core.NativeHandler
	}{
		"read_file": {
			Description: "Read a file from the local filesystem",
			Schema:      readFileSchema,
			Handler:     fs.ReadFile,
		},
		"write_file": {
			Description: "Write content to a file on the local filesystem",
			Schema:      writeFileSchema,
			Handler:     fs.WriteFile,
		},
		"edit_file": {
			Description: "Edit a file by replacing an exact string with a new one",
			Schema:      editFileSchema,
			Handler:     fs.EditFile,
		},
		"list_directory": {
			Description: "List contents of a directory",
			Schema:      listDirSchema,
			Handler:     fs.ListDir,
		},
		"search_files": {
			Description: "Search for a string in files recursively",
			Schema:      searchFilesSchema,
			Handler:     fs.SearchFiles,
		},
		"get_file_info": {
			Description: "Get metadata about a file (size, mode, modtime)",
			Schema:      getFileInfoSchema,
			Handler:     fs.GetFileInfo,
		},
		"glob": {
			Description: "Find files matching a pattern (supports ** for recursive search)",
			Schema:      globSchema,
			Handler:     fs.Glob,
		},
	}
}
