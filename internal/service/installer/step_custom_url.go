package installer

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type CustomURLStep struct {
	input textinput.Model
}

func NewCustomURLStep() Step {
	ti := textinput.New()
	ti.Focus()
	ti.Placeholder = "https://api.example.com/v1"
	ti.Width = 50
	return &CustomURLStep{input: ti}
}

func (s *CustomURLStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *CustomURLStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		val := strings.TrimSpace(s.input.Value())
		if val != "" {
			state.EnvVars["TUSK_CUSTOM_OPENAI_BASE_URL"] = val
			return nil, nil
		}
	}
	return s, cmd
}

func (s *CustomURLStep) View(state *InstallState) string {
	return "Enter Custom OpenAI Base URL:\n\n" + s.input.View() + "\n\n(press enter to confirm)\n"
}
