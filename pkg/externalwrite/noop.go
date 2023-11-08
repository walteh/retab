package externalwrite

import (
	"context"
	"io"
)

type NoopExternalFormatter struct {
}

func (me *NoopExternalFormatter) Format(ctx context.Context, reader io.Reader, writer io.Writer) func() error {
	return func() error {

		read := io.TeeReader(reader, writer)

	}
}
