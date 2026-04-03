package core

import "context"

type ReleaseInfo struct {
	Version string
	URL     string
	SHA256  string
}

type ReleaseProvider interface {
	GetLatestReleaseInfo(ctx context.Context) (*ReleaseInfo, error)
	GetReleaseBinary(ctx context.Context, info *ReleaseInfo) (string, error)
}

type Updater interface {
	Check(ctx context.Context) (*ReleaseInfo, error)
	Update(ctx context.Context, release *ReleaseInfo) error
}
