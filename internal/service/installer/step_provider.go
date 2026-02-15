package installer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ProviderStep allows selection of the AI provider
type ProviderStep struct {
	choices []string
	cursor  int
}

func NewProviderStep() Step {
	return &ProviderStep{
		choices: []string{"Anthropic", "OpenAI", "OpenRouter", "Ollama"},
		cursor:  0,
	}
}

func (s *ProviderStep) Init() tea.Cmd {
	return nil
}

func (s *ProviderStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}
		case "enter":
			state.EnvVars["TUSK_MODEL_PROVIDER"] = s.choices[s.cursor]
			return nil, nil
		}
	}
	return s, nil
}

func (s *ProviderStep) View(state *InstallState) string {
	var b strings.Builder
	b.WriteString("Select your AI Provider:\n\n")
	for i, choice := range s.choices {
		cursor := " "
		if s.cursor == i {
			cursor = "❯"
			b.WriteString(selStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n")
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n")
		}
	}
	b.WriteString("\n(press ctrl+c to quit)\n")
	return b.String()
}
