package installer

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	itemStyle  = lipgloss.NewStyle().PaddingLeft(2)
	selStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("5"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
)

// Step represents a single step in the installation wizard
type Step interface {
	Init() tea.Cmd
	Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd)
	View(state *InstallState) string
}

func getSteps() []Step {
	return []Step{
		NewProviderStep(),
		NewOpenRouterKeyStep(),
		NewModelStep(),
		NewDownloadModelStep(),
		NewChannelStep(),
		NewTelegramTokenStep(),
		NewTelegramOwnerStep(),
		NewFinalizationStep(),
		NewSaveEnvStep(),
		NewInitializeFilesStep(),
	}
}

type item struct {
	id    string
	title string
	desc  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.id }

type modelsMsg []list.Item
type errMsg error
type nextMsg struct{}

// model is the main Bubble Tea model that orchestrates the steps
type model struct {
	steps       []Step
	currentStep int
	state       *InstallState
	quitting    bool
	err         error
	width       int
	height      int
}

func initialModel() model {
	return model{
		steps:       getSteps(),
		currentStep: 0,
		state:       NewInstallState(),
	}
}

func (m model) Init() tea.Cmd {
	if len(m.steps) > 0 && m.steps[0] != nil {
		return m.steps[0].Init()
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case errMsg:
		m.err = msg
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	if m.currentStep >= len(m.steps) {
		return m, tea.Quit
	}

	nextStep, cmd := m.steps[m.currentStep].Update(msg, m.state, m.width, m.height)

	if nextStep == nil {
		// Step indicated completion, move to next
		m.currentStep++
		if m.currentStep >= len(m.steps) {
			// All steps completed
			return m, tea.Quit
		}
		// Initialize the next step
		return m, m.steps[m.currentStep].Init()
	}

	// If the step returned a different step (e.g., for branching), update current
	if nextStep != m.steps[m.currentStep] {
		m.steps[m.currentStep] = nextStep
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Installation cancelled.\n"
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n(press ctrl+c to quit)\n"
	}

	if m.currentStep >= len(m.steps) {
		return "Configuration complete!\n"
	}

	return titleStyle.Render("Installing TuskBot 🦣") + "\n\n" + m.steps[m.currentStep].View(m.state)
}

// RunWizard starts the TUI
func RunWizard() (*InstallState, error) {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	finalModel := m.(model)
	if finalModel.quitting {
		return nil, fmt.Errorf("tuskbot installation interrupted")
	}

	return finalModel.state, nil
}
