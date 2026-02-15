package installer

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type OllamaURLStep struct {
	input textinput.Model
}

func NewOllamaURLStep() Step {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "http://127.0.0.1:11434"
	return &OllamaURLStep{input: ti}
}

func (s *OllamaURLStep) Init() tea.Cmd { return nil }

func (s *OllamaURLStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	provider := strings.ToLower(state.EnvVars["TUSK_MODEL_PROVIDER"])
	if provider != "ollama" {
		return nil, nil
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		val := s.input.Value()
		if val == "" {
			val = s.input.Placeholder
		}
		state.EnvVars["TUSK_OLLAMA_BASE_URL"] = val
		return nil, nil
	}

	return s, cmd
}

func (s *OllamaURLStep) View(state *InstallState) string {
	return "Enter Ollama Base URL:\n\n" + s.input.View() + "\n\n(press enter to confirm)\n"
}
