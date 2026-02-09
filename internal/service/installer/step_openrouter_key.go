package installer

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// OpenRouterKeyStep collects the OpenRouter API key
type OpenRouterKeyStep struct {
	input textinput.Model
}

func NewOpenRouterKeyStep() Step {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 40
	ti.Placeholder = "sk-or-v1-..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &OpenRouterKeyStep{
		input: ti,
	}
}

func (s *OpenRouterKeyStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *OpenRouterKeyStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			state.EnvVars["TUSK_OPENROUTER_API_KEY"] = s.input.Value()
			return nil, nil
		}
	}
	return s, cmd
}

func (s *OpenRouterKeyStep) View(state *InstallState) string {
	return "Enter your OpenRouter API Key:\n\n" +
		s.input.View() + "\n\n" +
		"(press enter to confirm)\n"
}
