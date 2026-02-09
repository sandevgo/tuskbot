package installer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ChannelStep allows selection of the chat channel/transport
type ChannelStep struct {
	choices []string
	cursor  int
}

func NewChannelStep() Step {
	return &ChannelStep{
		choices: []string{"Telegram"},
		cursor:  0,
	}
}

func (s *ChannelStep) Init() tea.Cmd {
	return nil
}

func (s *ChannelStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
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
			state.EnvVars["TUSK_CHAT_CHANNEL"] = s.choices[s.cursor]
			return nil, nil
		}
	}
	return s, nil
}

func (s *ChannelStep) View(state *InstallState) string {
	var b strings.Builder
	b.WriteString("Select your Chat Channel:\n\n")
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
