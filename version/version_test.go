package version_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func TestVersionE2E(t *testing.T) {

	if os.Getenv("E2E") != "1" {
		t.SkipNow()
	}

	cmd := exec.Command("retab", "--version")
	cmd.Env = append([]string{}, os.Environ()...)
	// cmd.Dir = "/usr/bin"
	out, err := cmd.CombinedOutput()
	// defer func() {
	// 	if err != nil {
	// 		// run ls -l to make sure the binary is not stripped
	// 		ccmd := exec.Command("ls", "-l", "/usr/bin")
	// 		cmd.Env = append([]string{}, os.Environ()...)
	// 		o, err := ccmd.CombinedOutput()
	// 		require.NoError(t, err, string(o))
	// 		t.Log(string(o))
	// 	}
	// }()
	require.NoError(t, err, string(out))

	// There should be at least one newline and the first line
	// of output should contain the name, version, and possibly a revision.
	firstLine, _, hasNewline := strings.Cut(string(out), "\n")
	require.True(t, hasNewline, "At least one newline is required in the output")

	// Log the output to make debugging easier.
	t.Log(firstLine)

	// Split by spaces into at least 2 fields.
	fields := strings.Fields(firstLine)
	require.GreaterOrEqual(t, len(fields), 2, "Expected at least 2 fields in the first line, '%+v'", firstLine)

	// First field should be an import path.
	// This can be any valid import path for Go
	// so don't set too many restrictions here.
	// Just checking if the import path is a valid Go
	// path should be suitable enough to make sure this is ok.
	// Using CheckImportPath instead of CheckPath as it is less
	// restrictive.
	importPath := fields[0]
	require.NoError(t, module.CheckImportPath(importPath), "First field was not a valid import path: %+v", importPath)

	// Second field should be a version.
	// This defaults to something that's still compatible
	// with semver.
	version := fields[1]
	require.True(t, semver.IsValid(version), "Second field was not valid semver: %+v", version)

	// Revision should be empty or should look like a git hash.
	if len(fields) > 2 && len(fields[2]) > 0 {
		revision := fields[2]
		require.Regexp(t, `[0-9a-f]{40}`, revision, "Third field was not a git revision: %+v", revision)
	}
}
