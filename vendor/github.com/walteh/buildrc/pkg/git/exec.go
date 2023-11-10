package git

// var _ GitProvider = (*ExecGitProvider)(nil)

// type ExecGitProvider struct {
// }

// func NewExecGitProvider() *ExecGitProvider {
// 	return &ExecGitProvider{}
// }

// func (me *ExecGitProvider) GetContentHash(ctx context.Context) (string, error) {

// 	resolved, err := resolveCommitHash(ctx, "HEAD")
// 	if err != nil {
// 		return "", err
// 	}

// 	out, err := exec.Command("bash", "-c", fmt.Sprintf("git ls-tree -r %s | sha256sum", resolved)).Output()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Output is a byte slice, we convert it to string
// 	// Also, sha256sum outputs the hash followed by a '-', we only want the hash
// 	sha256sum := strings.Fields(string(out))[0]

// 	return sha256sum, nil
// }

// func resolveCommitHash(ctx context.Context, sha string) (string, error) {

// 	out, err := exec.Command("bash", "-c", fmt.Sprintf("git rev-parse %s", sha)).Output()
// 	if err != nil {
// 		return "", err
// 	}

// 	// Output is a byte slice, we convert it to string
// 	// Also, sha256sum outputs the hash followed by a '-', we only want the hash
// 	sha256sum := strings.Fields(string(out))[0]

// 	return sha256sum, nil
// }

// func (me *ExecGitProvider) GetCurrentCommitHash(ctx context.Context) (string, error) {
// 	return resolveCommitHash(ctx, "HEAD")
// }

// func (me *ExecGitProvider) GetCurrentBranch(ctx context.Context) (string, error) {
// 	out, err := exec.Command("bash", "-c", "git branch --show-current").Output()
// 	if err != nil {
// 		return "", err
// 	}

// 	return strings.TrimSpace(string(out)), nil
// }

// func (me *ExecGitProvider) GetLatestSemverTagFromRef(ctx context.Context, ref string) (*semver.Version, error) {

// 	resolved, err := resolveCommitHash(ctx, ref)
// 	if err != nil {
// 		return nil, err
// 	}

// 	cmd := exec.Command("git", "tag", "--merged", resolved)
// 	output, err := cmd.Output()
// 	if err != nil {
// 		return nil, errors.Errorf("failed to execute git command: %v", err)
// 	}

// 	// Parse the output
// 	tags := strings.Split(string(output), "\n")
// 	var versions []*semver.Version

// 	for _, tag := range tags {
// 		// Skip empty lines
// 		if len(tag) == 0 {
// 			continue
// 		}

// 		// Attempt to parse each tag as a semver version
// 		ver, err := semver.NewVersion(tag)
// 		if err != nil {
// 			return nil, errors.Errorf("failed to parse tag '%s' as semver: %v", tag, err)
// 		}
// 		versions = append(versions, ver)
// 	}

// 	// Return error if no semver tags found
// 	if len(versions) == 0 {
// 		zerolog.Ctx(ctx).Warn().Strs("tags", tags).Str("commit", resolved).Str("output", string(output)).Msg("no semver tags found")
// 		return nil, errors.Errorf("no semver tags found from ref '%s'", ref)
// 	}

// 	// Sort the versions in descending order
// 	sort.Sort(sort.Reverse(semver.Collection(versions)))

// 	// Return the latest version
// 	return versions[0], nil
// }

// func (me *ExecGitProvider) GetLocalRepositoryMetadata(ctx context.Context) (*LocalRepositoryMetadata, error) {
// 	panic("implement me")
// }
