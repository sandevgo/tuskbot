package updater

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Service struct {
	provider  core.ReleaseProvider
	systemSvc core.SystemService
	version   string
}

func NewService(provider core.ReleaseProvider, systemSvc core.SystemService, version string) *Service {
	return &Service{
		provider:  provider,
		systemSvc: systemSvc,
		version:   version,
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

	if isDevVersion(s.version) {
		return nil, nil
	}

	currentVersion := normalizeVersion(s.version)
	latestVersion := normalizeVersion(release.Version)
	if currentVersion != "" && latestVersion != "" {
		cmp, err := compareVersions(currentVersion, latestVersion)
		if err == nil && cmp >= 0 {
			return nil, nil
		}
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
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	return replaceExecutable(execPath, tmpPath)
}

func replaceExecutable(execPath, tmpPath string) error {
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

func isDevVersion(version string) bool {
	return strings.TrimSpace(strings.TrimPrefix(version, "v")) == "dev"
}

func compareVersions(currentVersion, latestVersion string) (int, error) {
	currentParts, err := parseVersion(currentVersion)
	if err != nil {
		return 0, err
	}

	latestParts, err := parseVersion(latestVersion)
	if err != nil {
		return 0, err
	}

	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		currentPart := versionPart(currentParts, i)
		latestPart := versionPart(latestParts, i)

		switch {
		case currentPart < latestPart:
			return -1, nil
		case currentPart > latestPart:
			return 1, nil
		}
	}

	return 0, nil
}

func parseVersion(version string) ([]int, error) {
	parts := strings.Split(version, ".")
	parsed := make([]int, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid version: %q", version)
		}

		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid version: %q", version)
		}
		parsed = append(parsed, n)
	}

	return parsed, nil
}

func versionPart(parts []int, idx int) int {
	if idx >= len(parts) {
		return 0
	}
	return parts[idx]
}
