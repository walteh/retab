package testutils

import (
	"os/exec"
	"strings"
	"testing"
)

func NewDisposableContainerName(t *testing.T) string {
	t.Helper()

	containerName := "retab_" + strings.ReplaceAll(t.Name(), "/", "_")

	t.Cleanup(func() {
		// remove the container
		exec.Command("docker", "rm", "-f", containerName).Run()
	})

	return containerName
}
