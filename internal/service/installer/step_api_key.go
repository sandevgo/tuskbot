package installer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// APIKeyStep collects provider-specific API keys (skips for Ollama if not needed)
type APIKeyStep struct {
	input      textinput.Model
	provider   string
	envKey     string
	title      string
	isOptional bool
}

func NewAPIKeyStep() Step {
	return &APIKeyStep{}
}

func (s *APIKeyStep) Init() tea.Cmd {
	return nil
}

func (s *APIKeyStep) initProvider(state *InstallState) bool {
	s.provider = strings.ToLower(state.EnvVars["TUSK_MODEL_PROVIDER"])
	if s.provider == "" {
		return false
	}

	switch s.provider {
	case "anthropic":
		s.envKey = "TUSK_ANTHROPIC_API_KEY"
		s.title = "Anthropic API Key"
	case "openai":
		s.envKey = "TUSK_OPENAI_API_KEY"
		s.title = "OpenAI API Key"
	case "openrouter":
		s.envKey = "TUSK_OPENROUTER_API_KEY"
		s.title = "OpenRouter API Key"
	case "ollama":
		s.envKey = "TUSK_OLLAMA_API_KEY"
		s.title = "Ollama API Key (Optional)"
		s.isOptional = true

		if state.EnvVars["TUSK_OLLAMA_BASE_URL"] == "" {
			state.EnvVars["TUSK_OLLAMA_BASE_URL"] = "http://localhost:11434"
		}
	default:
		return false
	}

	s.input = textinput.New()
	s.input.Focus()
	s.input.CharLimit = 255
	s.input.Width = 40
	s.input.EchoMode = textinput.EchoPassword
	s.input.EchoCharacter = '•'

	switch s.provider {
	case "anthropic":
		s.input.Placeholder = "sk-ant-..."
	case "openai":
		s.input.Placeholder = "sk-..."
	case "openrouter":
		s.input.Placeholder = "sk-or-v1-..."
	case "ollama":
		s.input.Placeholder = "Optional - press Enter to skip"
		s.input.EchoMode = textinput.EchoNormal
	}
	return true
}

func (s *APIKeyStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	if s.provider == "" {
		if !s.initProvider(state) {
			return nil, nil
		}
		return s, textinput.Blink
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			state.EnvVars[s.envKey] = s.input.Value()
			return nil, nil
		}
	}
	return s, cmd
}

func (s *APIKeyStep) View(state *InstallState) string {
	if s.provider == "" {
		if !s.initProvider(state) {
			return "Loading..."
		}
	}

	optionalHint := ""
	if s.isOptional {
		optionalHint = " (optional - press Enter to skip)"
	}
	
	return fmt.Sprintf("Enter your %s%s:\n\n%s\n\n(press enter to confirm)\n", 
		s.title, optionalHint, s.input.View())
}
