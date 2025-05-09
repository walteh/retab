//go:build !js

package cmdfmt

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"gitlab.com/tozd/go/errors"
)

func runBasicCmd(ctx context.Context, cmds []string) (string, error) {
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Errorf("failed to run command: %w %s", err, string(out))
	}

	return string(out), nil
}

func runFmtCmd(ctx context.Context, cmds []string, w io.Writer, r io.Reader, opts *BasicExternalFormatterOpts) error {

	fmt.Println("CMDS", strings.Join(cmds, " "))

	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = r
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Env = os.Environ()

	for cname, cdata := range opts.tempFiles {
		tempFile, err := os.CreateTemp("", cname)
		if err != nil {
			return errors.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tempFile.Name())
		_, err = tempFile.Write([]byte(cdata))
		if err != nil {
			return errors.Errorf("failed to write temp file: %w", err)
		}

		for i, arg := range cmd.Args {
			if arg == cname {
				cmd.Args[i] = tempFile.Name()
			}
		}
	}

	// out, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return errors.Errorf("failed to run command: %w", err)
	// }

	// fmt.Println("OUT", string(out))

	// read the from stdout

	return cmd.Run()
}
