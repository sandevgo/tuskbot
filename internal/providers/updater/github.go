package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type GitHubProvider struct {
	repo    string
	baseURL string
	client  *http.Client
}

func NewGitHubProvider(repo string) *GitHubProvider {
	return &GitHubProvider{
		repo:    repo,
		baseURL: "https://api.github.com",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *GitHubProvider) GetLatestReleaseInfo(ctx context.Context) (*core.ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", p.baseURL, p.repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "TuskBot-Updater")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api request failed: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode github response: %w", err)
	}

	assetName := releaseAssetName(runtime.GOOS, runtime.GOARCH)

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no release found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return &core.ReleaseInfo{
		Version: release.TagName,
		URL:     downloadURL,
	}, nil
}

func (p *GitHubProvider) GetReleaseBinary(ctx context.Context, info *core.ReleaseInfo) (string, error) {
	// 1. Download the archive
	req, err := http.NewRequestWithContext(ctx, "GET", info.URL, nil)
	if err != nil {
		return "", err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	// 2. Save to a temporary file (tar.gz)
	archiveFile, err := os.CreateTemp("", "tusk-update-*.tar.gz")
	if err != nil {
		return "", err
	}
	archivePath := archiveFile.Name()
	defer os.Remove(archivePath) // Clean up archive when done

	if _, err := io.Copy(archiveFile, resp.Body); err != nil {
		archiveFile.Close()
		return "", err
	}
	archiveFile.Close()

	// 3. Extract the binary from archive
	// The archive contains: bin/tusk-{os}-{arch}
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzf.Close()

	tarReader := tar.NewReader(gzf)
	allowedNames := releaseBinaryNames(runtime.GOOS, runtime.GOARCH)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar read error: %w", err)
		}

		// Look for the binary file. It's usually in a 'bin/' folder or at root.
		if header.Typeflag == tar.TypeReg {
			name := header.Name
			if !allowedNames[name] {
				continue
			}

			tmpPath, err := writeTempBinary(tarReader)
			if err != nil {
				return "", err
			}
			return tmpPath, nil
		}
	}

	return "", fmt.Errorf("binary not found inside archive")
}

func releaseAssetName(goos, goarch string) string {
	return fmt.Sprintf("tusk-%s-%s.tar.gz", goos, goarch)
}

func releaseBinaryNames(goos, goarch string) map[string]bool {
	name := fmt.Sprintf("tusk-%s-%s", goos, goarch)
	return map[string]bool{
		"bin/" + name: true,
		name:          true,
		"tusk":        true,
	}
}

func writeTempBinary(src io.Reader) (string, error) {
	tmpBinary, err := os.CreateTemp("", "tusk-binary-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmpBinary.Name()

	if _, err := io.Copy(tmpBinary, src); err != nil {
		tmpBinary.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to read binary content: %w", err)
	}

	if err := tmpBinary.Chmod(0755); err != nil {
		tmpBinary.Close()
		os.Remove(tmpPath)
		return "", err
	}

	if err := tmpBinary.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}

	return tmpPath, nil
}
