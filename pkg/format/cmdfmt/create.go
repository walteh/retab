//go:build !js

package cmdfmt

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

func NewExecFormatter(ctx context.Context, opts *BasicExternalFormatterOpts, cmds ...string) format.Provider {
	return ExternalFormatterToProvider(&basicExternalFormatter{opts.Indent, opts.TempFiles, func(r io.Reader, w io.Writer) func() error {
		if len(cmds) < 1 {
			return func() error {
				return errors.New("no command specified")
			}
		}

		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdin = r
		cmd.Stdout = w
		cmd.Stderr = w
		cmd.Env = os.Environ()

		return func() error {
			for cname, cdata := range opts.TempFiles {
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

			return cmd.Run()
		}
	}})
}
