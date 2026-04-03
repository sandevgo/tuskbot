package updater

import (
	"context"
	"os"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Service struct {
	provider  core.ReleaseProvider
	systemSvc core.SystemService
	version   string
	replaceFn func(string) error
}

func NewService(provider core.ReleaseProvider, systemSvc core.SystemService, version string) *Service {
	return &Service{
		provider:  provider,
		systemSvc: systemSvc,
		version:   version,
		replaceFn: nil,
	}
}

func (s *Service) Check(ctx context.Context) (*core.ReleaseInfo, error) {
	release, err := s.provider.GetLatestReleaseInfo(ctx)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, nil
	}

	currentVersion := normalizeVersion(s.version)
	latestVersion := normalizeVersion(release.Version)
	if currentVersion != "" && currentVersion == latestVersion {
		return nil, nil
	}
	return release, nil
}

func (s *Service) Update(ctx context.Context, release *core.ReleaseInfo) error {
	shouldRestart, err := s.stopSystemService(ctx)
	if err != nil {
		return err
	}

	if shouldRestart {
		defer s.systemSvc.Start(ctx)
	}

	tmpPath, err := s.provider.GetReleaseBinary(ctx, release)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	return s.replace(tmpPath)
}

func (s *Service) stopSystemService(ctx context.Context) (bool, error) {
	if s.systemSvc != nil {
		status, err := s.systemSvc.Status(ctx)
		if err != nil {
			return false, err
		}

		if status == core.SystemServiceStatusRunning {
			if err := s.systemSvc.Stop(ctx); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) replace(tmpPath string) error {
	if s.replaceFn != nil {
		return s.replaceFn(tmpPath)
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	oldPath := execPath + ".old"
	_ = os.Remove(oldPath)

	if err := os.Rename(execPath, oldPath); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		_ = os.Rename(oldPath, execPath)
		return err
	}

	_ = os.Remove(oldPath)
	return nil
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	if version == "" || version == "dev" {
		return ""
	}
	return version
}
