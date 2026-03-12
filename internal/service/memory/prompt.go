package memory

import (
	"fmt"
	"os"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type SysPrompt struct {
	path string
	cfg  core.PromptConfig
}

func NewSysPrompt(runtimePath string, cfg core.PromptConfig) *SysPrompt {
	return &SysPrompt{
		path: runtimePath,
		cfg:  cfg,
	}
}

func (p *SysPrompt) BuildForAgent() []core.Message {
	messages := make([]core.Message, 0)

	if content := readFile(p.cfg.GetSystemPath()); content != "" {
		now := time.Now().Format("2006-01-02 15:04:05 GMT-07")
		sysPrompt := fmt.Sprintf(content, now, p.path)
		messages = append(messages, core.Message{Role: "system", Content: sysPrompt})
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

func (p *SysPrompt) BuildForSubAgent() []core.Message {
	messages := make([]core.Message, 0)
	if content := readFile(p.cfg.GetSubAgentPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	return messages
}

func readFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}
