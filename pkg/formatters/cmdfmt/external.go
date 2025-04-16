package cmdfmt

import (
	"context"
	"io"

	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type ExternalFormatterConfig struct {
	Indentation string
	Targets     []string
}

type ExternalFormatter interface {
	Format(ctx context.Context, reader io.Reader) (io.Reader, func() error)
	Indent() string
	TempFiles() map[string]string
}

type externalStdinFormatter struct {
	internal ExternalFormatter
}

func ExternalFormatterToProvider(ext ExternalFormatter) format.Provider {
	return &externalStdinFormatter{ext}
}

func (me *externalStdinFormatter) Format(ctx context.Context, cfg format.Configuration, input io.Reader) (io.Reader, error) {

	read, f := me.internal.Format(ctx, input)

	var rerr error
	go func() {
		if err := f(); err != nil {
			rerr = err
		}
	}()

	output, err := format.BruteForceIndentation(ctx, me.internal.Indent(), cfg, read)
	if err != nil {
		return nil, errors.Errorf("failed to apply configuration: %w", err)
	}

	if rerr != nil {
		return nil, errors.Errorf("failed to format: %w", rerr)
	}

	return output, nil
}

// type externalFileFormatter struct {
// 	internal ExternalFileFormatter
// 	fmter    ExternalFormatter
// }

// func (me *externalFileFormatter) Targets() []string {
// 	return me.internal.Targets()
// }

// func ExternalFileFormatterToProvider(ext ExternalFormatter) format.Provider {
// 	return &externalStdinFormatter{ext}
// }
