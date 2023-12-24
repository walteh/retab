package externalwrite

import (
	"context"
	"io"
	"os/exec"

	"github.com/go-faster/errors"
	"github.com/rs/zerolog"

	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/terrors"
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
				return terrors.New("no command specified")
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
				return errors.Wrap(err, "failed to copy")
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
				return terrors.Wrap(err, "failed to close pipe")
			}
			return terrors.Wrap(err, "failed to run command")
		}
		if err := pipw.Close(); err != nil {
			return terrors.Wrap(err, "failed to close pipe")
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
