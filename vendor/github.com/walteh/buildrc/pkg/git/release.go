package git

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

type Release struct {
	ID         string
	CommitHash string
	Tag        string
	PR         *PullRequest
	Artifacts  []string
	Draft      bool
}

type ReleaseProvider interface {
	UploadReleaseArtifact(ctx context.Context, id string, name string, file afero.File) error
	DownloadReleaseArtifact(ctx context.Context, id string, name string, filesystem afero.Fs) (afero.File, error)
	DeleteReleaseArtifact(ctx context.Context, id string, name string) error
	HasReleaseArtifact(ctx context.Context, id string, name string) (bool, error)
	GetReleaseByTag(ctx context.Context, tag string) (*Release, error)
	GetReleaseByID(ctx context.Context, id string) (*Release, error)
	TagRelease(ctx context.Context, prov GitProvider, vers *semver.Version) (*Release, error)
	ListRecentReleases(ctx context.Context, limit int) ([]*Release, error)
	TakeReleaseOutOfDraft(ctx context.Context, id string) error
}

func ReleaseAlreadyExists(ctx context.Context, prov ReleaseProvider, gitp GitProvider) (bool, string, error) {

	current, err := gitp.GetCurrentCommitFromRef(ctx, "HEAD")
	if err != nil {
		return false, "", err
	}

	releases, err := prov.ListRecentReleases(ctx, 100)
	if err != nil {
		return false, "", err
	}

	for _, rel := range releases {
		if current == rel.CommitHash && !rel.Draft {
			zerolog.Ctx(ctx).Info().Str("tag", rel.Tag).Any("release", rel).Msg("release already exists")
			return true, rel.Tag, nil
		}
	}

	return false, "", nil
}

func (me *Release) Semver() (*semver.Version, error) {
	return semver.NewVersion(me.Tag)
}
