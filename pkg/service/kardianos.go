package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	kservice "github.com/kardianos/service"
	"github.com/sandevgo/tuskbot/internal/core"
)

type KardianosManager struct {
	svc kservice.Service
}

var _ core.SystemService = (*KardianosManager)(nil)

func NewKardianosManager(cfg Config) (*KardianosManager, error) {
	options := kservice.KeyValue{
		"WorkingDirectory": cfg.WorkingDirectory,
		"UserService":      cfg.UserService,
	}
	if cfg.SandboxMode {
		options["SystemdScript"] = systemdSandboxScript
		options["LaunchdConfig"] = launchdSandboxConfig
	}

	svcCfg := &kservice.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
		Option:      options,
	}
	if cfg.Executable != "" {
		svcCfg.Executable = cfg.Executable
	}

	svc, err := kservice.New(nil, svcCfg)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &KardianosManager{svc: svc}, nil
}

func (m *KardianosManager) Install(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status != core.SystemServiceStatusUnknown {
		return nil
	}

	if err := m.svc.Install(); err != nil {
		if isAlreadyExistsErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Uninstall(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status == core.SystemServiceStatusUnknown {
		return nil
	}

	if err := m.svc.Uninstall(); err != nil {
		if isNotInstalledErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status == core.SystemServiceStatusRunning {
		return nil
	}

	if err := m.svc.Start(); err != nil {
		if isAlreadyRunningErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status != core.SystemServiceStatusRunning {
		return nil
	}

	if err := m.svc.Stop(); err != nil {
		if isNotRunningErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Status(ctx context.Context) (core.SystemServiceStatus, error) {
	if err := ctx.Err(); err != nil {
		return core.SystemServiceStatusUnknown, err
	}

	st, err := m.svc.Status()
	if err != nil {
		if isNotInstalledErr(err) {
			return core.SystemServiceStatusUnknown, nil
		}
		return core.SystemServiceStatusUnknown, err
	}

	switch st {
	case kservice.StatusRunning:
		return core.SystemServiceStatusRunning, nil
	case kservice.StatusStopped:
		return core.SystemServiceStatusStopped, nil
	default:
		return core.SystemServiceStatusUnknown, nil
	}
}

func isAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "already exists") || strings.Contains(e, "already installed") || strings.Contains(e, "service exists")
}

func isNotInstalledErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, kservice.ErrNotInstalled) {
		return true
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "not installed") || strings.Contains(e, "does not exist") || strings.Contains(e, "could not find")
}

func isAlreadyRunningErr(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "already started") || strings.Contains(e, "already running") || strings.Contains(e, "service is running")
}

func isNotRunningErr(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "not loaded") || strings.Contains(e, "not running") || strings.Contains(e, "not started") || strings.Contains(e, "input/output error")
}
