package installer

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TelegramTokenStep collects the Telegram bot token
type TelegramTokenStep struct {
	input textinput.Model
}

func NewTelegramTokenStep() Step {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 40
	ti.Placeholder = "123456789:ABCDEF..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &TelegramTokenStep{
		input: ti,
	}
}

func (s *TelegramTokenStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *TelegramTokenStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			state.EnvVars["TUSK_TELEGRAM_TOKEN"] = s.input.Value()
			return nil, nil
		}
	}
	return s, cmd
}

func (s *TelegramTokenStep) View(state *InstallState) string {
	return "Enter your Telegram Bot Token:\n\n" +
		s.input.View() + "\n\n" +
		"(press enter to confirm)\n"
}

// TelegramOwnerStep collects the Telegram owner ID
type TelegramOwnerStep struct {
	input textinput.Model
}

func NewTelegramOwnerStep() Step {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	ti.Width = 40
	ti.Placeholder = "123456789"
	ti.EchoMode = textinput.EchoNormal

	return &TelegramOwnerStep{
		input: ti,
	}
}

func (s *TelegramOwnerStep) Init() tea.Cmd {
	return textinput.Blink
}

func (s *TelegramOwnerStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			state.EnvVars["TUSK_TELEGRAM_OWNER_ID"] = s.input.Value()
			// Wizard completed successfully
			return nil, nil
		}
	}
	return s, cmd
}

func (s *TelegramOwnerStep) View(state *InstallState) string {
	return "Enter your Telegram User ID (Owner):\n\n" +
		s.input.View() + "\n\n" +
		"(press enter to confirm)\n"
}
