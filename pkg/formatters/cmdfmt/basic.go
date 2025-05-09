package cmdfmt

import (
	"context"
	"io"

	"github.com/rs/zerolog"
	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type basicExternalFormatter struct {
	indent    string
	tempFiles map[string]string
	f         func(io.Reader, io.Writer) func(ctx context.Context) error
}

//go:opts
type BasicExternalFormatterOpts struct {
	indent    string
	tempFiles map[string]string

	useDocker bool

	dockerImageName string
	dockerImageTag  string
	dockerCommand   []string
	fmtCommand      []string
	executable      string
}

func NewNoopBasicExternalFormatProvider() format.Provider {
	return WrapExternalFormatterWithStdio(&basicExternalFormatter{"  ", map[string]string{}, func(r io.Reader, w io.Writer) func(ctx context.Context) error {
		return func(ctx context.Context) error {
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
		if err := cmd(ctx); err != nil {
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

// TempFiles implements format.ExternalFormatter.
func (me *basicExternalFormatter) TempFiles() map[string]string {
	return me.tempFiles
}
