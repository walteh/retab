package tests

import (
	"os"
	"os/exec"
)

type cmdOpt func(*exec.Cmd)

func withArgs(args ...string) cmdOpt {
	return func(cmd *exec.Cmd) {
		cmd.Args = append(cmd.Args, args...)
	}
}

func mainCmd(opts ...cmdOpt) *exec.Cmd {
	cmd := exec.Command("tftab")
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}
