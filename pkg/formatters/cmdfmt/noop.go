package cmdfmt

import (
	"context"
	"io"

	"github.com/walteh/retab/v2/pkg/format"
)

type NoopExternalFormatter struct {
}

var _ ExternalFormatter = (*NoopExternalFormatter)(nil)

func NewNoopExternalFormatProvider() format.Provider {
	return WrapExternalFormatterWithStdio(&NoopExternalFormatter{})
}
func (me *NoopExternalFormatter) Format(_ context.Context, input io.Reader) (io.Reader, func() error) {
	return input, func() error {
		return nil
	}
}

func (me *NoopExternalFormatter) Indent() string {
	return "  "
}

func (me *NoopExternalFormatter) TempFiles() map[string]string {
	return map[string]string{}
}
