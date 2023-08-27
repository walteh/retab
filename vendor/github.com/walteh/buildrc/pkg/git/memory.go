package git

import (
	"context"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/afero"
)

/* /////////////////////////////////////////
	REPOMETADATA PROVIDER
///////////////////////////////////////// */

type memoryRepoMetadataProvider struct {
	cmd *RemoteRepositoryMetadata
}

func NewMemoryRepoMetadataProvider(cmd *RemoteRepositoryMetadata) RemoteRepositoryMetadataProvider {
	return &memoryRepoMetadataProvider{cmd: cmd}
}

func (me *memoryRepoMetadataProvider) GetRemoteRepositoryMetadata(ctx context.Context) (*RemoteRepositoryMetadata, error) {
	return me.cmd, nil
}

/* /////////////////////////////////////////
	PULLREQUEST PROVIDER
///////////////////////////////////////// */

type memoryPullRequestProvider struct {
	prs []*PullRequest
}

func NewMemoryPullRequestProvider(prs []*PullRequest) PullRequestProvider {
	return &memoryPullRequestProvider{prs: prs}
}

func (me *memoryPullRequestProvider) ListRecentPullRequests(ctx context.Context, head string) ([]*PullRequest, error) {
	return me.prs, nil
}

/* /////////////////////////////////////////
	RELEASE PROVIDER
///////////////////////////////////////// */

type memoryReleaseProvider struct {
	rels []*Release
}

func NewMemoryReleaseProvider(rels []*Release) ReleaseProvider {
	return &memoryReleaseProvider{rels: rels}
}

func (me *memoryReleaseProvider) UploadReleaseArtifact(ctx context.Context, id string, name string, file afero.File) error {
	// r.Artifacts = append(r.Artifacts, name)
	return nil
}

func (me *memoryReleaseProvider) DownloadReleaseArtifact(ctx context.Context, id string, name string, filesystem afero.Fs) (afero.File, error) {
	return filesystem.Create(name)
}

func (me *memoryReleaseProvider) GetReleaseByTag(ctx context.Context, tag string) (*Release, error) {
	for _, r := range me.rels {
		if r.Tag == tag {
			return r, nil
		}
	}
	return nil, nil
}

func (me *memoryReleaseProvider) GetReleaseByID(ctx context.Context, id string) (*Release, error) {
	for _, r := range me.rels {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}

func (me *memoryReleaseProvider) TagRelease(ctx context.Context, r GitProvider, vers *semver.Version) (*Release, error) {
	return &Release{
		Tag:        "v" + vers.String(),
		CommitHash: "1234567890",
	}, nil
}

func (me *memoryReleaseProvider) ListRecentReleases(ctx context.Context, limit int) ([]*Release, error) {
	return me.rels, nil
}

func (me *memoryReleaseProvider) DeleteReleaseArtifact(ctx context.Context, id string, name string) error {
	return nil
}

func (me *memoryReleaseProvider) HasReleaseArtifact(ctx context.Context, id string, name string) (bool, error) {
	return false, nil
}

func (me *memoryReleaseProvider) TakeReleaseOutOfDraft(ctx context.Context, id string) error {
	return nil
}
