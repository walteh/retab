package buildrc

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-faster/errors"

	"github.com/rs/zerolog"
	"github.com/walteh/buildrc/pkg/git"
)

type CommitType string

const (
	CommitTypePR      CommitType = "pr"
	CommitTypeLocal   CommitType = "local"
	CommitTypeRelease CommitType = "release"
)

type GetVersionOpts struct {
	Type                  CommitType `json:"type"`
	PatchIndicator        string     `json:"patch-indicator"`
	PRNumber              uint64     `json:"pr-number"`
	CommitMessageOverride string     `json:"commit-message-override"`
	LatestTagOverride     string     `json:"latest-tag-override"`
	Patch                 bool       `json:"patch"`
	Auto                  bool       `json:"auto"`
	ExcludeV              bool       `json:"exclude-v"`
}

func GetVersion(ctx context.Context, gitp git.GitProvider, brc *Buildrc, me *GetVersionOpts) (string, error) {

	zerolog.Ctx(ctx).Debug().Any("buildrc", brc).Msg("loading buildrc file")

	if me == nil {
		me = &GetVersionOpts{Auto: true, PatchIndicator: "patch"}
	}

	prefix := "v"
	if me.ExcludeV {
		prefix = ""
	}

	if me.Patch {
		me.PatchIndicator = "patch"
		me.CommitMessageOverride = "patch"
	}

	if me.Type == CommitTypePR {
		if me.PRNumber == 0 {
			return "", errors.Errorf("'--pr-number=#' is required for type %s", me.Type)
		}
	}

	if me.Auto {
		me.Type = CommitTypeRelease
		if gitp.Dirty(ctx) {
			me.Type = CommitTypeLocal
		} else {
			svt, err := gitp.TryGetSemverTag(ctx)
			if err != nil {
				return "", err
			}

			if svt != nil {
				return prefix + svt.String(), nil
			}

			n, err := gitp.TryGetPRNumber(ctx)
			if err != nil {
				return "", err
			}

			me.PRNumber = n
			if me.PRNumber > 0 {
				me.Type = CommitTypePR
			}
		}
	}

	switch me.Type {
	case CommitTypeRelease:
		{

			var latestHead *semver.Version
			var message string
			var err error

			if me.LatestTagOverride != "" {
				latestHead, err = semver.NewVersion(me.LatestTagOverride)
				if err != nil {
					return "", err
				}
			} else {
				latestHead, err = gitp.GetLatestSemverTagFromRef(ctx, "HEAD")
				if err != nil {
					return "", err
				}
			}

			if me.CommitMessageOverride != "" {
				message = me.CommitMessageOverride
			} else {
				message, err = gitp.GetCurrentCommitMessageFromRef(ctx, "HEAD")
				if err != nil {
					return "", err
				}
			}

			patch := strings.Contains(message, me.PatchIndicator)

			if latestHead.Major() < brc.Major() {
				latestHead, err = semver.NewVersion(strconv.FormatUint(brc.Major(), 10) + ".0.0")
				if err != nil {
					return "", err
				}
				return prefix + latestHead.String(), nil
			}

			// we do not care about the prerelease or metadata and this safely removes it
			work := *semver.New(latestHead.Major(), latestHead.Minor(), latestHead.Patch(), "", "")

			if patch {
				work = work.IncPatch()
			} else {
				work = work.IncMinor()
			}

			return prefix + work.String(), nil

		}
	case CommitTypeLocal:
		{
			work := *semver.New(0, 0, 0, "local", time.Now().Format("2006.01.02.15.04.05"))
			return prefix + work.String(), nil
		}
	case CommitTypePR:
		{

			latestHead, err := gitp.GetLatestSemverTagFromRef(ctx, "HEAD")
			if err != nil {
				return "", err
			}

			if latestHead.Major() < brc.Major() {
				latestHead, err = semver.NewVersion(strconv.FormatUint(brc.Major(), 10) + ".0.0")
				if err != nil {
					return "", err
				}
			}

			revision, err := gitp.GetCurrentShortHashFromRef(ctx, "HEAD")
			if err != nil {
				return "", err
			}

			work := *latestHead

			work, err = work.SetPrerelease("pr." + strconv.FormatUint(me.PRNumber, 10))
			if err != nil {
				return "", err
			}

			work, err = work.SetMetadata(revision)
			if err != nil {
				return "", err
			}

			return prefix + work.String(), nil
		}
	}

	return "", errors.Errorf("unknown type %s", me.Type)
}
