package installer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/providers/llm"
)

// ModelStep allows selection of the AI model from OpenRouter
type ModelStep struct {
	list     list.Model
	loading  bool
	fetching bool // Ensures we only trigger the API call once
	err      error
}

func NewModelStep() Step {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select AI Model"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return &ModelStep{
		list:    l,
		loading: true,
	}
}

func (s *ModelStep) Init() tea.Cmd {
	return nil
}

func (s *ModelStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	// 1. Trigger fetch once when we enter the step
	if s.loading && !s.fetching {
		s.fetching = true
		apiKey := state.EnvVars["TUSK_OPENROUTER_API_KEY"]

		return s, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cfg := &config.AppConfig{
				OpenRouterAPIKey: apiKey,
			}
			p := llm.NewOpenRouter(cfg)
			models, err := p.GetModels(ctx)
			if err != nil {
				return errMsg(err)
			}

			var items []list.Item
			for _, mod := range models {
				items = append(items, item{
					id:    mod.ID,
					title: mod.Name,
					desc:  fmt.Sprintf("ID: %s | Context: %d", mod.ID, mod.ContextLength),
				})
			}
			return modelsMsg(items)
		}
	}

	// Update list size
	s.list.SetSize(width, height-4)

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case modelsMsg:
		s.list.SetItems(msg)
		s.loading = false
		s.fetching = false
		return s, nil

	case errMsg:
		s.loading = false
		s.fetching = false
		s.err = msg
		return s, nil // Return nil command to break the error loop

	case tea.KeyMsg:
		// If there's an error, allow retry with Enter
		if s.err != nil {
			if msg.String() == "enter" {
				s.err = nil
				s.loading = true
				s.fetching = false
				return s, nil
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
				provider := state.EnvVars["TUSK_MODEL_PROVIDER"]
				state.EnvVars["TUSK_MAIN_MODEL"] = fmt.Sprintf("%s/%s", strings.ToLower(provider), i.id)
				return nil, nil
			}
			return s, cmd
		}
	}

	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

func (s *ModelStep) View(state *InstallState) string {
	if s.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error fetching models: %v", s.err)) +
			"\n\nCheck your API key and internet connection.\n\n(press enter to retry, ctrl+c to quit)\n"
	}
	if s.loading {
		return "Fetching models from OpenRouter...\n"
	}
	return s.list.View()
}
