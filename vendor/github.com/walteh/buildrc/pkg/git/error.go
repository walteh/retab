package git

import "fmt"

type GitError error

var (
	ErrNoGitProvider GitError = GitError(fmt.Errorf("no git provider found"))
	ErrNoMatchingPR  GitError = GitError(fmt.Errorf("no matching PR found"))
	ErrRefNotFound   GitError = GitError(fmt.Errorf("ref not found"))
)
