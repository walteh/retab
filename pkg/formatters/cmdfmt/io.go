package cmdfmt

import (
	"context"
	"io"

	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type externalStdioFormatter struct {
	internal ExternalFormatter
}

func WrapExternalFormatterWithStdio(ext ExternalFormatter) format.Provider {
	return &externalStdioFormatter{ext}
}

func (me *externalStdioFormatter) Format(ctx context.Context, cfg format.Configuration, input io.Reader) (io.Reader, error) {

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
