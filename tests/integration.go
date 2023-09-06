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

	// run ls -la
	cmd1 := exec.Command("/bin/ls", "-la", "/usr/bin/")
	cmd1.Env = append([]string{}, os.Environ()...)
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr

	defer func() {
		if err := cmd1.Run(); err != nil {
			panic(err)
		}
	}()

	cmd := exec.Command("/usr/bin/retab")
	cmd.Env = append([]string{}, os.Environ()...)
	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
}
