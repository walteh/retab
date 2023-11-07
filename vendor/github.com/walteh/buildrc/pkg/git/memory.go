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

func (me *memoryRepoMetadataProvider) GetRemoteRepositoryMetadata(_ context.Context) (*RemoteRepositoryMetadata, error) {
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

func (me *memoryPullRequestProvider) ListRecentPullRequests(_ context.Context, _ string) ([]*PullRequest, error) {
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

func (me *memoryReleaseProvider) UploadReleaseArtifact(_ context.Context, _ string, _ string, _ afero.File) error {
	// r.Artifacts = append(r.Artifacts, name)
	return nil
}

func (me *memoryReleaseProvider) DownloadReleaseArtifact(_ context.Context, _ string, name string, filesystem afero.Fs) (afero.File, error) {
	return filesystem.Create(name)
}

func (me *memoryReleaseProvider) GetReleaseByTag(_ context.Context, tag string) (*Release, error) {
	for _, r := range me.rels {
		if r.Tag == tag {
			return r, nil
		}
	}
	return nil, nil
}

func (me *memoryReleaseProvider) GetReleaseByID(_ context.Context, id string) (*Release, error) {
	for _, r := range me.rels {
		if r.ID == id {
			return r, nil
		}
	}
	return nil, nil
}

func (me *memoryReleaseProvider) TagRelease(_ context.Context, _ GitProvider, vers *semver.Version) (*Release, error) {
	return &Release{
		Tag:        "v" + vers.String(),
		CommitHash: "1234567890",
	}, nil
}

func (me *memoryReleaseProvider) ListRecentReleases(_ context.Context, _ int) ([]*Release, error) {
	return me.rels, nil
}

func (me *memoryReleaseProvider) DeleteReleaseArtifact(_ context.Context, _ string, _ string) error {
	return nil
}

func (me *memoryReleaseProvider) HasReleaseArtifact(_ context.Context, _ string, _ string) (bool, error) {
	return false, nil
}

func (me *memoryReleaseProvider) TakeReleaseOutOfDraft(_ context.Context, _ string) error {
	return nil
}
