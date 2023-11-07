package file

import (
	"context"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/afero"
)

func FilterGitIgnored(_ context.Context, fls afero.Fs, lines []string) ([]string, error) {
	ignoreFile, err := afero.ReadFile(fls, ".gitignore")
	if err != nil {
		return nil, err
	}

	strs := strings.Split(string(ignoreFile), "\n")
	strs = append(strs, ".git")

	igns := ignore.CompileIgnoreLines(strs...)
	if err != nil {
		return nil, err
	}

	filtered := []string{}
	for _, line := range lines {
		liner := strings.Split(line, " ")
		lined := liner[len(liner)-1]
		if !igns.MatchesPath(lined) {
			filtered = append(filtered, line)
		}
	}

	return filtered, nil

}
