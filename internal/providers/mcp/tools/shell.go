package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const executeCommandSchema = `
{
  "type": "object",
  "properties": {
    "command": { "type": "string", "description": "The shell command to execute" }
  },
  "required": ["command"]
}
`

const (
	maxOutputLines     = 200
	defaultExecTimeout = 5 * time.Minute
)

type Shell struct {
	WorkDir string
}

func NewShell(workDir string) *Shell {
	return &Shell{WorkDir: workDir}
}

func (s *Shell) ExecuteCommand(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	// Create a child context with a timeout to prevent hanging commands
	ctx, cancel := context.WithTimeout(ctx, defaultExecTimeout)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", input.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", input.Command)
	}

	if s.WorkDir != "" {
		cmd.Dir = s.WorkDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := s.truncateOutput(stdout.String())
	errOutput := s.truncateOutput(stderr.String())

	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Command timed out after %v\nSTDOUT:\n%s\nSTDERR:\n%s", defaultExecTimeout, output, errOutput), nil
		}
		return fmt.Sprintf("Command failed: %v\nSTDOUT:\n%s\nSTDERR:\n%s", err, output, errOutput), nil
	}

	return fmt.Sprintf("STDOUT:\n%s\nSTDERR:\n%s", output, errOutput), nil
}

func (s *Shell) truncateOutput(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return "(empty)"
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxOutputLines {
		return output
	}

	truncated := lines[len(lines)-maxOutputLines:]
	return fmt.Sprintf("... (output truncated, showing last %d lines)\n%s", maxOutputLines, strings.Join(truncated, "\n"))
}

func (s *Shell) GetDefinitions() map[string]struct {
	Description string
	Schema      string
	Handler     func(context.Context, json.RawMessage) (string, error)
} {
	return map[string]struct {
		Description string
		Schema      string
		Handler     func(context.Context, json.RawMessage) (string, error)
	}{
		"execute_command": {"Execute a shell command", executeCommandSchema, s.ExecuteCommand},
	}
}
