package mcp

import "time"

type Timeouts struct {
	Connect  time.Duration
	ToolList time.Duration
	ToolCall time.Duration
}

func NewDefaultTimeouts() *Timeouts {
	return &Timeouts{
		Connect:  30 * time.Second,
		ToolList: 5 * time.Second,
		ToolCall: 2 * time.Minute,
	}
}
