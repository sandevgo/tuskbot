package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sandevgo/tuskbot/internal/core"
)

func TestGitHubProviderGetLatestReleaseInfo(t *testing.T) {
	t.Parallel()

	assetName := releaseAssetName(runtime.GOOS, runtime.GOARCH)

	tests := []struct {
		name       string
		repo       string
		statusCode int
		body       string
		wantErr    bool
		wantURL    string
	}{
		{
			name:       "selects platform asset using repo slug path",
			repo:       "sandevgo/tuskbot",
			statusCode: http.StatusOK,
			body: `{
				"tag_name":"v1.2.3",
				"assets":[
					{"name":"other.tar.gz","browser_download_url":"https://example.com/other.tar.gz"},
					{"name":"` + assetName + `","browser_download_url":"https://example.com/` + assetName + `"}
				]
			}`,
			wantURL: "https://example.com/" + assetName,
		},
		{
			name:       "returns error when platform asset is missing",
			repo:       "sandevgo/tuskbot",
			statusCode: http.StatusOK,
			body: `{
				"tag_name":"v1.2.3",
				"assets":[
					{"name":"wrong.tar.gz","browser_download_url":"https://example.com/wrong.tar.gz"}
				]
			}`,
			wantErr: true,
		},
		{
			name:       "returns error on github api failure",
			repo:       "sandevgo/tuskbot",
			statusCode: http.StatusForbidden,
			body:       `{}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotPath string
			provider := NewGitHubProvider(tt.repo)
			provider.baseURL = "https://example.test"
			provider.client = &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					gotPath = r.URL.Path
					return newResponse(tt.statusCode, tt.body), nil
				}),
			}

			release, err := provider.GetLatestReleaseInfo(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetLatestReleaseInfo() error = %v", err)
			}
			if gotPath != "/repos/"+tt.repo+"/releases/latest" {
				t.Fatalf("request path = %q, want %q", gotPath, "/repos/"+tt.repo+"/releases/latest")
			}
			if release == nil {
				t.Fatal("expected release info, got nil")
			}
			if release.Version != "v1.2.3" {
				t.Fatalf("release.Version = %q, want %q", release.Version, "v1.2.3")
			}
			if release.URL != tt.wantURL {
				t.Fatalf("release.URL = %q, want %q", release.URL, tt.wantURL)
			}
		})
	}
}

func TestGitHubProviderGetReleaseBinary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		entries     map[string]string
		wantContent string
		wantErr     bool
	}{
		{
			name: "prefers bin target",
			entries: map[string]string{
				"README.md": "docs",
				"bin/tusk-" + runtime.GOOS + "-" + runtime.GOARCH: "bin-binary",
				"tusk": "fallback",
			},
			wantContent: "bin-binary",
		},
		{
			name: "accepts root target",
			entries: map[string]string{
				"tusk-" + runtime.GOOS + "-" + runtime.GOARCH: "root-binary",
			},
			wantContent: "root-binary",
		},
		{
			name: "accepts plain tusk fallback",
			entries: map[string]string{
				"tusk": "plain-binary",
			},
			wantContent: "plain-binary",
		},
		{
			name: "ignores lookalike files",
			entries: map[string]string{
				"docs/tusk-notes.txt": "notes",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			archive := buildArchive(t, tt.entries)
			provider := NewGitHubProvider(core.TuskRepositorySlug)
			provider.client = &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					return newResponse(http.StatusOK, string(archive)), nil
				}),
			}

			path, err := provider.GetReleaseBinary(context.Background(), &core.ReleaseInfo{URL: "https://example.test/download"})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetReleaseBinary() error = %v", err)
			}
			defer os.Remove(path)

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", path, err)
			}
			if string(data) != tt.wantContent {
				t.Fatalf("binary content = %q, want %q", string(data), tt.wantContent)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("Stat(%q) error = %v", path, err)
			}
			if info.Mode().Perm() != 0o755 {
				t.Fatalf("binary mode = %v, want %v", info.Mode().Perm(), os.FileMode(0o755))
			}
		})
	}
}

func buildArchive(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var archive bytes.Buffer
	gzw := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gzw)

	for name, content := range entries {
		hdr := &tar.Header{
			Name: filepath.ToSlash(name),
			Mode: 0o755,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("WriteHeader(%q) error = %v", name, err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("tar close error = %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("gzip close error = %v", err)
	}

	return archive.Bytes()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func newResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     strings.TrimSpace(http.StatusText(statusCode)),
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header:     make(http.Header),
	}
}
