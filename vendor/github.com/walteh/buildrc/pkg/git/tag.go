package git

import (
	"context"
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.com/rs/zerolog"
)

type TagStragegy string

const (
	TagStrategyCommitToMain       TagStragegy = "commit-to-main"
	TagStrategySquashMerge        TagStragegy = "squash-merge"
	TagStrategyMerge              TagStragegy = "merge"
	TagStrategyCommitToExistingPR TagStragegy = "commit-to-existing-pr"
	TagStrategyCommitToNewPR      TagStragegy = "commit-to-new-pr"
)

func CalculateTagStrategy(ctx context.Context, git GitProvider, prp PullRequestProvider) (TagStragegy, *semver.Version, *PullRequest, error) {

	latestHead, err := git.GetLatestSemverTagFromRef(ctx, "HEAD")
	if err != nil {
		return "", nil, nil, err
	}

	latestMain, err := git.GetLatestSemverTagFromRef(ctx, "main")
	if err != nil {
		return "", nil, nil, err
	}

	var highest *semver.Version
	if latestHead.GreaterThan(latestMain) {
		highest = latestHead
	} else {
		highest = latestMain
	}

	pr, err := getLatestPullRequestForRef(ctx, prp, "HEAD")
	if err != nil {
		return "", nil, nil, err
	}

	// if there is no pr, then this was a direct commit to main
	// so we just increment the patch version

	brnch, err := git.GetCurrentBranchFromRef(ctx, "HEAD")
	if err != nil {
		return "", nil, nil, err
	}

	if brnch != "main" {
		if pr != nil {
			if pr.Closed {
				return "", nil, nil, errors.New("pr is closed - please create a new pr")
			}

			if latestHead.Equal(latestMain) {
				// this is a new pr
				return TagStrategyCommitToNewPR, latestMain, pr, nil
			}

			return TagStrategyCommitToExistingPR, highest, pr, nil
		} else {
			return "", nil, nil, errors.New("no pr found - please create a pr")
		}
	}

	pr, err = getLatestMergedPullRequestThatHasAMatchingContentHash(ctx, prp, git)
	if err != nil {
		if errors.Is(err, ErrNoMatchingPR) {
			// then this is a direct commit to main
			// so we just increment the patch version
			return TagStrategyCommitToMain, highest, pr, nil
		}
		return "", nil, nil, err
	}

	svr, err := git.GetLatestSemverTagFromRef(ctx, pr.Head)
	if err != nil {
		return "", nil, nil, err
	}

	if svr.Equal(latestMain) {
		// then this is a merge commit
		return TagStrategyMerge, svr, pr, nil
	}

	// if there is a pr, then this was a squash merge
	return TagStrategySquashMerge, svr, pr, nil

}

func CalculateNextPreReleaseTag(ctx context.Context, brc uint64, git GitProvider, prp PullRequestProvider) (*semver.Version, error) {

	strat, last, pr, err := CalculateTagStrategy(ctx, git, prp)
	if err != nil {
		return nil, err
	}

	brcv := semver.New(uint64(brc), 0, 0, "", "")

	if last.LessThan(brcv) {
		last = brcv
	}

	zerolog.Ctx(ctx).Debug().Str("strategy", string(strat)).Str("last", last.String()).Msg("calculated tag strategy")

	cmt, err := git.GetCurrentShortHashFromRef(ctx, "HEAD")
	if err != nil {
		return nil, err
	}

	switch strat {
	case TagStrategyCommitToMain:
		strt := last.IncPatch()
		return &strt, nil
	case TagStrategySquashMerge, TagStrategyMerge:
		return semver.New(last.Major(), last.Minor(), last.Patch(), "", ""), nil
	case TagStrategyCommitToExistingPR:
		strt, err := last.SetMetadata(cmt)
		if err != nil {
			return nil, err
		}
		return &strt, nil
	case TagStrategyCommitToNewPR:
		if pr == nil {
			return nil, errors.New("no pr found in commit to new pr strategy")
		}
		strt := last.IncMinor()
		strt, err = strt.SetPrerelease(pr.PreReleaseTag())
		if err != nil {
			return nil, err
		}
		strt, err = strt.SetMetadata(cmt)
		if err != nil {
			return nil, err
		}
		return &strt, nil
	default:
		return nil, errors.New("unknown tag strategy")
	}
}

// func CalculateNextPreReleaseTags(ctx context.Context, brc *buildrc.Buildrc, git GitProvider, prp PullRequestProvider) (*semver.Version, error) {

// 	latestHead, err := git.GetLatestSemverTagFromRef(ctx, "HEAD")
// 	if err != nil {
// 		return nil, err
// 	}

// 	latestMain, err := git.GetLatestSemverTagFromRef(ctx, "main")
// 	if err != nil {
// 		return nil, err
// 	}

// 	latestMajor := semver.New(uint64(brc.Version), 0, 0, "", "")

// 	pr, err := getLatestOpenOrMergedPullRequestForRef(ctx, prp, "HEAD")
// 	if err != nil {
// 		return nil, err
// 	}

// 	if pr == nil {
// 		// if there is no pr, then this was a direct commit to main
// 		// so we just increment the patch version

// 		brnch, err := git.GetCurrentBranchFromRef(ctx, "HEAD")
// 		if err != nil {
// 			return nil, err
// 		}

// 		if brnch != "main" {
// 			return nil, errors.New("no pr found - please create a pr")
// 		}

// 		pr, err = getLatestMergedPullRequestThatHasAMatchingContentHash(ctx, prp, git)
// 		if err != nil {
// 			if errors.Is(err, ErrNoMatchingPR) {
// 				// then this is a direct commit to main
// 				// so we just increment the patch version
// 				if latestMain == nil {
// 					latestMain = latestMajor
// 				}
// 				res := latestMain.IncPatch()
// 				return &res, nil
// 			}
// 			return nil, err
// 		}

// 		tag, err := git.GetLatestSemverTagFromRef(ctx, pr.Head)
// 		if err != nil {
// 			return nil, err
// 		}

// 		return semver.New(tag.Major(), tag.Minor(), tag.Patch(), "", ""), nil
// 	}

// 	prefix := pr.PreReleaseTag()

// 	if latestMain != nil && latestMain.GreaterThan(latestHead) {
// 		latestHead = latestMain
// 	}

// 	if latestMajor.GreaterThan(latestHead) {
// 		latestHead = latestMajor
// 	}

// 	shouldInc := !strings.Contains(latestHead.Prerelease(), prefix)

// 	var result semver.Version

// 	if shouldInc {
// 		result = latestHead.IncMinor()
// 	} else {
// 		result = *latestHead
// 	}

// 	result, err = result.SetPrerelease(prefix)
// 	if err != nil {
// 		return nil, err
// 	}

// 	zerolog.Ctx(ctx).Debug().Str("prefix", prefix).Any("latestHead", latestHead).Any("result", result).Msg("release version")

// 	return &result, nil
// }
