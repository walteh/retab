package cmdfmt

import (
	"context"
	"io"

	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type ExternalFormatter interface {
	Format(ctx context.Context, reader io.Reader) (io.Reader, func() error)
	Indent() string
	TempFiles() map[string]string
}

func NewFormatter(cmds []string, optz ...OptBasicExternalFormatterOptsSetter) format.Provider {
	opts := NewBasicExternalFormatterOpts(optz...)

	if opts.executable == "" {
		panic("executable is empty")
	}

	if opts.useDocker {
		return NewDockerCmdFormatter(cmds, optz...)
	}

	cmds = append([]string{opts.executable}, cmds...)

	return NewCmdFormatter(cmds, optz...)
}

func NewCmdFormatter(cmds []string, optz ...OptBasicExternalFormatterOptsSetter) format.Provider {
	opts := NewBasicExternalFormatterOpts(optz...)

	basic := &basicExternalFormatter{
		indent:    opts.indent,
		tempFiles: opts.tempFiles,
		f: func(r io.Reader, w io.Writer) func(ctx context.Context) error {
			if len(cmds) < 1 {
				return func(ctx context.Context) error {
					return errors.New("no command specified")
				}
			}

			return func(ctx context.Context) error {
				return runFmtCmd(ctx, cmds, w, r, &opts)
			}
		}}

	return WrapExternalFormatterWithStdio(basic)
}
