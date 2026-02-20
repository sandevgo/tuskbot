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

func (fs *Filesystem) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     func(context.Context, json.RawMessage) (string, error)
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}{
		"read_file":      {"Read a file from the local filesystem", readFileSchema, fs.ReadFile},
		"write_file":     {"Write content to a file on the local filesystem", writeFileSchema, fs.WriteFile},
		"edit_file":      {"Edit a file by replacing an exact string with a new one", editFileSchema, fs.EditFile},
		"list_directory": {"List contents of a directory", listDirSchema, fs.ListDir},
		"search_files":   {"Search for a string in files recursively", searchFilesSchema, fs.SearchFiles},
		"get_file_info":  {"Get metadata about a file (size, mode, modtime)", getFileInfoSchema, fs.GetFileInfo},
	}
}
