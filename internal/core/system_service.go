package core

import "context"

type SystemServiceStatus string

const (
	SystemServiceStatusUnknown SystemServiceStatus = "unknown"
	SystemServiceStatusRunning SystemServiceStatus = "running"
	SystemServiceStatusStopped SystemServiceStatus = "stopped"
)

type SystemService interface {
	Install(ctx context.Context) error
	Uninstall(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status(ctx context.Context) (SystemServiceStatus, error)
}
