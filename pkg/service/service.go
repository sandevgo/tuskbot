package service

import "context"

type Status string

const (
	StatusUnknown Status = "unknown"
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
)

type Config struct {
	Name             string
	DisplayName      string
	Description      string
	Executable       string
	Arguments        []string
	WorkingDirectory string
}

type Manager interface {
	Install(ctx context.Context) error
	Uninstall(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (Status, error)
}
