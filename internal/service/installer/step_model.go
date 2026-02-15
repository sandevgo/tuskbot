package installer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/llm"
)

type modelInitMsg struct{}

// ModelStep allows selection of the AI model based on provider
type ModelStep struct {
	list       list.Model
	loading    bool
	fetching   bool
	err        error
	provider   string
	manualMode bool
	input      textinput.Model
}

func NewModelStep() Step {
	return &ModelStep{}
}

func (s *ModelStep) Init() tea.Cmd {
	return func() tea.Msg { return modelInitMsg{} }
}

func (s *ModelStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	if s.provider == "" {
		s.provider = strings.ToLower(state.EnvVars["TUSK_MODEL_PROVIDER"])
	}

	if s.list.Title == "" && s.provider != "" {
		l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
		l.Title = fmt.Sprintf("Select %s Model", strings.Title(s.provider))
		l.SetShowStatusBar(true)
		l.SetFilteringEnabled(true)
		l.Styles.Title = titleStyle
		s.list = l

		s.loading = true
		s.fetching = true
		return s, s.fetchModels(state)
	}

	if s.manualMode {
		var cmd tea.Cmd
		s.input, cmd = s.input.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				modelName := s.input.Value()
				state.EnvVars["TUSK_MAIN_MODEL"] = fmt.Sprintf("%s/%s", s.provider, modelName)
				return nil, nil
			}
		}
		return s, cmd
	}

	if !s.manualMode && s.provider != "" {
		s.list.SetSize(width, height-4)
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case modelsMsg:
		s.list.SetItems(msg)
		s.loading = false
		s.fetching = false
		return s, nil

	case errMsg:
		if s.provider == "ollama" {
			s.manualMode = true
			s.loading = false
			s.fetching = false
			s.err = nil

			s.input = textinput.New()
			s.input.Focus()
			s.input.CharLimit = 100
			s.input.Width = 40
			s.input.Placeholder = "llama3.2, mistral, codellama..."
			return s, textinput.Blink
		}
		s.loading = false
		s.fetching = false
		s.err = msg
		return s, nil

	case tea.KeyMsg:
		if s.err != nil {
			if msg.String() == "enter" {
				s.err = nil
				s.loading = true
				return s, s.fetchModels(state)
			}
			return s, nil
		}

		if msg.String() == "enter" {
			wasFiltering := s.list.FilterState() == list.Filtering
			s.list, cmd = s.list.Update(msg)

			if wasFiltering || s.list.FilterState() == list.Filtering {
				return s, cmd
			}

			if i, ok := s.list.SelectedItem().(item); ok {
				state.EnvVars["TUSK_MAIN_MODEL"] = fmt.Sprintf("%s/%s", s.provider, i.id)
				return nil, nil
			}
			return s, cmd
		}
	}

	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s *ModelStep) fetchModels(state *InstallState) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var provider core.AIProvider

		switch s.provider {
		case "openai":
			apiKey := state.EnvVars["TUSK_OPENAI_API_KEY"]
			provider = llm.NewOpenAI(apiKey, "")
		case "anthropic":
			apiKey := state.EnvVars["TUSK_ANTHROPIC_API_KEY"]
			provider = llm.NewAnthropic(apiKey, "")
		case "openrouter":
			apiKey := state.EnvVars["TUSK_OPENROUTER_API_KEY"]
			provider = llm.NewOpenRouter(apiKey, "")
		case "ollama":
			baseURL := state.EnvVars["TUSK_OLLAMA_BASE_URL"]
			apiKey := state.EnvVars["TUSK_OLLAMA_API_KEY"]
			provider = llm.NewOllama(baseURL, apiKey, "")
		default:
			return errMsg(fmt.Errorf("unknown provider: %s", s.provider))
		}

		models, err := provider.Models(ctx)
		if err != nil {
			return errMsg(err)
		}

		items := make([]list.Item, 0, len(models))
		for _, m := range models {
			items = append(items, item{
				id:    m.ID,
				title: m.Name,
				desc:  fmt.Sprintf("ID: %s | Context: %d", m.ID, m.ContextLength),
			})
		}
		return modelsMsg(items)
	}
}

func (s *ModelStep) View(state *InstallState) string {
	if s.provider == "" && !s.manualMode {
		return "Loading..."
	}

	if s.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error fetching models: %v", s.err)) +
			"\n\n(press enter to retry, ctrl+c to quit)\n"
	}

	if s.manualMode {
		return "Ollama connection failed. Enter model name manually:\n\n" +
			s.input.View() + "\n\n" +
			"Examples: llama3.2, mistral, codellama, phi3\n" +
			"(press enter to confirm)\n"
	}

	if s.loading {
		return fmt.Sprintf("Fetching available models from %s...\n", s.provider)
	}

	return s.list.View()
}
