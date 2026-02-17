package installer

import (
	tea "github.com/charmbracelet/bubbletea"
)

// FinalizationStep computes derived values and final env var formatting
type FinalizationStep struct{}

func NewFinalizationStep() Step {
	return &FinalizationStep{}
}

func (s *FinalizationStep) Init() tea.Cmd {
	return func() tea.Msg { return nextMsg{} }
}

func (s *FinalizationStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	// Set derived values
	if state.EnvVars["TUSK_TELEGRAM_TOKEN"] != "" {
		state.EnvVars["TUSK_ENABLE_TELEGRAM"] = "true"
	} else {
		state.EnvVars["TUSK_ENABLE_TELEGRAM"] = "false"
	}

	// Set defaults
	if state.EnvVars["TUSK_DEBUG"] == "" {
		state.EnvVars["TUSK_DEBUG"] = "0"
	}

	// Only used as intermediate state
	delete(state.EnvVars, "TUSK_MODEL_PROVIDER")

	// Signal completion
	return nil, nil
}

func (s *FinalizationStep) View(state *InstallState) string {
	return "Finalizing configuration...\n"
}
