package git

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
)

type GitProvider interface {
	LocalRepositoryMetadataProvider
	GetCurrentShortHashFromRef(ctx context.Context, ref string) (string, error)
	GetCurrentCommitFromRef(ctx context.Context, ref string) (string, error)
	GetCurrentCommitMessageFromRef(ctx context.Context, ref string) (string, error)
	GetCurrentBranchFromRef(ctx context.Context, ref string) (string, error)
	GetLatestSemverTagFromRef(ctx context.Context, ref string) (*semver.Version, error)
	GetContentHashFromRef(ctx context.Context, ref string) (string, error)
	TryGetPRNumber(ctx context.Context) (uint64, error)
	TryGetSemverTag(ctx context.Context) (*semver.Version, error)
	GetRemoteURL(ctx context.Context) (string, error)
	Dirty(ctx context.Context) bool

	Fs() afero.Fs
}
