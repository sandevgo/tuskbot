package memory

import (
	"os"

	"github.com/sandevgo/tuskbot/internal/core"
)

type SysPrompt struct {
	cfg core.PromptConfig
}

func NewSysPrompt(cfg core.PromptConfig) *SysPrompt {
	return &SysPrompt{
		cfg: cfg,
	}
}

func (p *SysPrompt) Build() []core.Message {
	messages := make([]core.Message, 0)
	readFile := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		return string(content)
	}

	if content := readFile(p.cfg.GetSystemPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	if content := readFile(p.cfg.GetIdentityPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	if content := readFile(p.cfg.GetUserProfilePath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	if content := readFile(p.cfg.GetMemoryPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	return messages
}
