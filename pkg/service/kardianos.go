package service

import (
	"context"
	"fmt"

	kservice "github.com/kardianos/service"
)

type kardianosManager struct {
	svc kservice.Service
}

func NewKardianosManager(cfg Config) (Manager, error) {
	svcCfg := &kservice.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
		Option: kservice.KeyValue{
			"WorkingDirectory": cfg.WorkingDirectory,
		},
	}
	if cfg.Executable != "" {
		svcCfg.Executable = cfg.Executable
	}

	svc, err := kservice.New(nil, svcCfg)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &kardianosManager{svc: svc}, nil
}

func (m *kardianosManager) Install(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.svc.Install()
}

func (m *kardianosManager) Uninstall(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.svc.Uninstall()
}

func (m *kardianosManager) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.svc.Start()
}

func (m *kardianosManager) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return m.svc.Stop()
}

func (m *kardianosManager) Status(ctx context.Context) (Status, error) {
	if err := ctx.Err(); err != nil {
		return StatusUnknown, err
	}

	st, err := m.svc.Status()
	if err != nil {
		return StatusUnknown, err
	}

	switch st {
	case kservice.StatusRunning:
		return StatusRunning, nil
	case kservice.StatusStopped:
		return StatusStopped, nil
	default:
		return StatusUnknown, nil
	}
}
