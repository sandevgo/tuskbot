package updater

import (
	"context"
	"errors"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

func TestServiceCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		currentVersion string
		release        *core.ReleaseInfo
		wantNil        bool
	}{
		{
			name:           "same normalized version returns nil",
			currentVersion: "1.2.3",
			release:        &core.ReleaseInfo{Version: "v1.2.3"},
			wantNil:        true,
		},
		{
			name:           "dev version does not update to released version",
			currentVersion: "dev",
			release:        &core.ReleaseInfo{Version: "v1.2.3"},
			wantNil:        true,
		},
		{
			name:           "different version returns release",
			currentVersion: "1.2.2",
			release:        &core.ReleaseInfo{Version: "v1.2.3"},
		},
		{
			name:           "older release does not downgrade current binary",
			currentVersion: "0.9.0",
			release:        &core.ReleaseInfo{Version: "v0.8.0"},
			wantNil:        true,
		},
		{
			name:           "same version with missing patch is treated as equal",
			currentVersion: "1.2",
			release:        &core.ReleaseInfo{Version: "v1.2.0"},
			wantNil:        true,
		},
		{
			name:           "invalid release version falls back to update available",
			currentVersion: "1.2.3",
			release:        &core.ReleaseInfo{Version: "latest"},
		},
		{
			name:           "missing release returns nil",
			currentVersion: "1.2.3",
			release:        nil,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewService(stubReleaseProvider{release: tt.release}, nil, tt.currentVersion)
			got, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Fatalf("Check() = %+v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("Check() returned nil release")
			}
			if got.Version != tt.release.Version {
				t.Fatalf("Check().Version = %q, want %q", got.Version, tt.release.Version)
			}
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		status           core.SystemServiceStatus
		statusErr        error
		stopErr          error
		downloadErr      error
		replaceErr       error
		wantErr          bool
		wantStopCalls    int
		wantStartCalls   int
		wantReplaceCalls int
	}{
		{
			name:             "running service stops and restarts",
			status:           core.SystemServiceStatusRunning,
			wantStopCalls:    1,
			wantStartCalls:   1,
			wantReplaceCalls: 1,
		},
		{
			name:             "stopped service does not restart",
			status:           core.SystemServiceStatusStopped,
			wantReplaceCalls: 1,
		},
		{
			name:      "status error aborts update",
			statusErr: errors.New("status failed"),
			wantErr:   true,
		},
		{
			name:          "stop error aborts update",
			status:        core.SystemServiceStatusRunning,
			stopErr:       errors.New("stop failed"),
			wantErr:       true,
			wantStopCalls: 1,
		},
		{
			name:        "download error does not restart stopped service",
			status:      core.SystemServiceStatusStopped,
			downloadErr: errors.New("download failed"),
			wantErr:     true,
		},
		{
			name:             "replace error still restarts previously running service",
			status:           core.SystemServiceStatusRunning,
			replaceErr:       errors.New("replace failed"),
			wantErr:          true,
			wantStopCalls:    1,
			wantStartCalls:   1,
			wantReplaceCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sys := &stubSystemService{
				status:    tt.status,
				statusErr: tt.statusErr,
				stopErr:   tt.stopErr,
			}
			provider := stubReleaseProvider{
				binaryPath: t.TempDir() + "/tusk.tmp",
				err:        tt.downloadErr,
			}
			svc := NewService(provider, sys, "1.0.0")
			svc.replaceFn = func(path string) error {
				if path == "" {
					t.Fatal("replace called with empty path")
				}
				sys.replaceCalls++
				return tt.replaceErr
			}

			err := svc.Update(context.Background(), &core.ReleaseInfo{Version: "v1.1.0"})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("Update() error = %v", err)
			}

			if sys.stopCalls != tt.wantStopCalls {
				t.Fatalf("stopCalls = %d, want %d", sys.stopCalls, tt.wantStopCalls)
			}
			if sys.startCalls != tt.wantStartCalls {
				t.Fatalf("startCalls = %d, want %d", sys.startCalls, tt.wantStartCalls)
			}
			if sys.replaceCalls != tt.wantReplaceCalls {
				t.Fatalf("replaceCalls = %d, want %d", sys.replaceCalls, tt.wantReplaceCalls)
			}
		})
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{input: "v1.2.3", want: "1.2.3"},
		{input: " 1.2.3 ", want: "1.2.3"},
		{input: "dev", want: ""},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := normalizeVersion(tt.input); got != tt.want {
				t.Fatalf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current string
		latest  string
		want    int
		wantErr bool
	}{
		{
			name:    "latest is newer",
			current: "1.2.3",
			latest:  "1.2.4",
			want:    -1,
		},
		{
			name:    "current is newer",
			current: "0.9.0",
			latest:  "0.8.0",
			want:    1,
		},
		{
			name:    "missing patch compares equal",
			current: "1.2",
			latest:  "1.2.0",
			want:    0,
		},
		{
			name:    "invalid version returns error",
			current: "1.2.3",
			latest:  "latest",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := compareVersions(tt.current, tt.latest)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("compareVersions() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("compareVersions() = %d, want %d", got, tt.want)
			}
		})
	}
}

type stubReleaseProvider struct {
	release    *core.ReleaseInfo
	binaryPath string
	err        error
}

func (s stubReleaseProvider) GetLatestReleaseInfo(context.Context) (*core.ReleaseInfo, error) {
	return s.release, s.err
}

func (s stubReleaseProvider) GetReleaseBinary(context.Context, *core.ReleaseInfo) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.binaryPath, nil
}

type stubSystemService struct {
	status       core.SystemServiceStatus
	statusErr    error
	stopErr      error
	startErr     error
	stopCalls    int
	startCalls   int
	replaceCalls int
}

func (s *stubSystemService) Install(context.Context) error {
	return nil
}

func (s *stubSystemService) Uninstall(context.Context) error {
	return nil
}

func (s *stubSystemService) Start(context.Context) error {
	s.startCalls++
	return s.startErr
}

func (s *stubSystemService) Stop(context.Context) error {
	s.stopCalls++
	return s.stopErr
}

func (s *stubSystemService) Status(context.Context) (core.SystemServiceStatus, error) {
	if s.statusErr != nil {
		return core.SystemServiceStatusUnknown, s.statusErr
	}
	return s.status, nil
}
