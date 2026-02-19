package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/providers/rag"
)

type progressMsg float64
type downloadDoneMsg string

type DownloadModelStep struct {
	progress progress.Model
	updates  chan tea.Msg
	err      error
	done     bool
	path     string
}

func NewDownloadModelStep() Step {
	return &DownloadModelStep{
		progress: progress.New(progress.WithDefaultGradient()),
		updates:  make(chan tea.Msg),
	}
}

func (s *DownloadModelStep) Init() tea.Cmd {
	// Start download in background
	go s.doDownload()
	// Start listening for updates
	return s.waitForActivity()
}

func (s *DownloadModelStep) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		return <-s.updates
	}
}

func (s *DownloadModelStep) doDownload() {
	runtimePath := config.GetRuntimePath()
	modelsDir := filepath.Join(runtimePath, "models")
	modelName := rag.ModelNameE5BaseQ8
	destPath := filepath.Join(modelsDir, modelName)

	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		s.updates <- errMsg(err)
		return
	}

	// Skip if exists
	if _, err := os.Stat(destPath); err == nil {
		s.updates <- downloadDoneMsg(modelName)
		return
	}

	resp, err := http.Get(rag.ModelUrlE5BaseQ8)
	if err != nil {
		s.updates <- errMsg(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.updates <- errMsg(fmt.Errorf("server returned status %d", resp.StatusCode))
		return
	}

	out, err := os.Create(destPath)
	if err != nil {
		s.updates <- errMsg(err)
		return
	}
	defer out.Close()

	reader := &progressReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
		onProgress: func(p float64) {
			s.updates <- progressMsg(p)
		},
	}

	if _, err := io.Copy(out, reader); err != nil {
		s.updates <- errMsg(err)
		return
	}

	s.updates <- downloadDoneMsg(modelName)
}

func (s *DownloadModelStep) Update(msg tea.Msg, state *InstallState, width, height int) (Step, tea.Cmd) {
	s.progress.Width = width - 10

	switch msg := msg.(type) {
	case progressMsg:
		var cmds []tea.Cmd
		cmds = append(cmds, s.waitForActivity())
		cmds = append(cmds, s.progress.SetPercent(float64(msg)))
		return s, tea.Batch(cmds...)

	case downloadDoneMsg:
		state.EnvVars["TUSK_EMBEDDING_MODEL"] = string(msg)
		s.done = true
		s.path = string(msg)
		return nil, nil

	case errMsg:
		s.err = msg
		return s, nil

	case progress.FrameMsg:
		progressModel, cmd := s.progress.Update(msg)
		s.progress = progressModel.(progress.Model)
		return s, cmd

	case tea.WindowSizeMsg:
		s.progress.Width = msg.Width - 10
	}

	return s, nil
}

func (s *DownloadModelStep) View(state *InstallState) string {
	if s.err != nil {
		return errorStyle.Render(fmt.Sprintf("Download failed: %v", s.err)) + "\n\n(press ctrl+c to quit)\n"
	}
	if s.done {
		return fmt.Sprintf("Model ready at: %s\n", s.path)
	}

	return "Downloading Embedding Model (GGUF)...\nThis may take a few minutes depending on your connection.\n\n" +
		s.progress.View() + "\n"
}

type progressReader struct {
	io.Reader
	Total      int64
	Downloaded int64
	onProgress func(float64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Downloaded += int64(n)
	if pr.Total > 0 {
		pr.onProgress(float64(pr.Downloaded) / float64(pr.Total))
	}
	return n, err
}
