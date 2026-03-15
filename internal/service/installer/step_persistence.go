package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
