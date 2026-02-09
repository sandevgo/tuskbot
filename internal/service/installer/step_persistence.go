package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	fs "github.com/sandevgo/tuskbot/configs"
	"github.com/sandevgo/tuskbot/internal/config"
)

// SaveEnvStep writes the collected configuration to .env file
type SaveEnvStep struct {
	err   error
	saved bool
}

func NewSaveEnvStep() Step {
	return &SaveEnvStep{}
}

func (s *SaveEnvStep) Init() tea.Cmd {
	return func() tea.Msg { return nextMsg{} }
}

func (s *SaveEnvStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	if s.saved {
		return nil, nil
	}

	// Perform save synchronously (fast operation)
	path := config.GetRuntimePath()

	if err := os.MkdirAll(path, 0755); err != nil {
		s.err = fmt.Errorf("failed to create runtime directory: %w", err)
		return s, nil
	}

	envPath := filepath.Join(path, ".env")

	// Check if .env already exists
	if _, err := os.Stat(envPath); err == nil {
		s.err = fmt.Errorf(".env file already exists at %s", envPath)
		return s, nil
	}

	// Build content from envVars map
	var content strings.Builder
	for key, value := range state.EnvVars {
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	if err := os.WriteFile(envPath, []byte(content.String()), 0600); err != nil {
		s.err = err
		return s, nil
	}

	s.saved = true
	return nil, nil // Signal completion
}

func (s *SaveEnvStep) View(state *InstallState) string {
	if s.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", s.err)) + "\n\n(press ctrl+c to quit)\n"
	}
	if s.saved {
		return "Configuration saved successfully!\n"
	}
	return "Saving configuration...\n"
}

// InitializeFilesStep writes the embedded configuration files to the runtime directory
type InitializeFilesStep struct {
	err  error
	done bool
}

func NewInitializeFilesStep() Step {
	return &InitializeFilesStep{}
}

func (s *InitializeFilesStep) Init() tea.Cmd {
	return nil
}

func (s *InitializeFilesStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	if s.done {
		return nil, nil
	}

	path := config.GetRuntimePath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		s.err = fmt.Errorf("failed to create runtime directory: %w", err)
		return s, nil
	}

	files := map[string]string{
		"IDENTITY.md":     filepath.Join(path, "IDENTITY.md"),
		"MEMORY.md":       filepath.Join(path, "MEMORY.md"),
		"SYSTEM.md":       filepath.Join(path, "SYSTEM.md"),
		"USER.md":         filepath.Join(path, "USER.md"),
		"mcp_config.json": filepath.Join(path, "mcp_config.json"),
	}

	for src, dst := range files {
		data, err := fs.FS.ReadFile(src)
		if err != nil {
			s.err = fmt.Errorf("failed to read embedded %s: %w", src, err)
			return s, nil
		}

		// Inject runtime path into SYSTEM.md
		if src == "SYSTEM.md" {
			content := fmt.Sprintf(string(data), path)
			data = []byte(content)
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			s.err = fmt.Errorf("failed to write %s: %w", dst, err)
			return s, nil
		}
	}

	s.done = true
	return nil, nil
}

func (s *InitializeFilesStep) View(state *InstallState) string {
	if s.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", s.err)) + "\n\n(press ctrl+c to quit)\n"
	}
	if s.done {
		return "Runtime files initialized successfully!\n"
	}
	return "Initializing runtime files...\n"
}
