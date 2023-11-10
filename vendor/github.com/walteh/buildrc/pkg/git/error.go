package git

import "github.com/go-faster/errors"

type GitError error

var (
	ErrNoGitProvider GitError = GitError(errors.Errorf("no git provider found"))
	ErrNoMatchingPR  GitError = GitError(errors.Errorf("no matching PR found"))
	ErrRefNotFound   GitError = GitError(errors.Errorf("ref not found"))
)
