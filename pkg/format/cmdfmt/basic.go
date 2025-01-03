package cmdfmt

import (
	"context"
	"io"
	"os/exec"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"

	"github.com/walteh/retab/v2/pkg/format"
)

type basicExternalFormatter struct {
	indent  string
	targets []string
	f       func(io.Reader, io.Writer) func() error
}

type BasicExternalFormatterOpts struct {
	Indent  string
	Targets []string
}

func NewExecFormatter(opts *BasicExternalFormatterOpts, cmds ...string) format.Provider {
	return ExternalFormatterToProvider(&basicExternalFormatter{opts.Indent, opts.Targets, func(r io.Reader, w io.Writer) func() error {
		if len(cmds) < 1 {
			return func() error {
				return errors.New("no command specified")
			}
		}
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdin = r
		cmd.Stdout = w
		cmd.Stderr = w
		return cmd.Run
	}})
}

func NewNoopBasicExternalFormatProvider() format.Provider {
	return ExternalFormatterToProvider(&basicExternalFormatter{"  ", []string{"*"}, func(r io.Reader, w io.Writer) func() error {
		return func() error {
			_, err := io.Copy(w, r)
			if err != nil {
				return errors.Errorf("failed to copy: %w", err)
			}
			return nil
		}
	}})
}

var _ ExternalFormatter = (*basicExternalFormatter)(nil)

// Format implements format.ExternalFormatter.
func (me *basicExternalFormatter) Format(ctx context.Context, reader io.Reader) (io.Reader, func() error) {
	zerolog.Ctx(ctx).Debug().Msg("running external formatter")
	pipr, pipw := io.Pipe()
	cmd := me.f(reader, pipw)
	return pipr, func() error {
		if err := cmd(); err != nil {
			err := pipw.CloseWithError(err)
			if err != nil {
				return errors.Errorf("failed to close pipe: %w", err)
			}
			return errors.Errorf("failed to run command: %w", err)
		}
		if err := pipw.Close(); err != nil {
			return errors.Errorf("failed to close pipe: %w", err)
		}
		return nil
	}
}

// Indent implements format.ExternalFormatter.
func (me *basicExternalFormatter) Indent() string {
	return me.indent
}

func (me *basicExternalFormatter) Targets() []string {
	return me.targets
}
