package externalwrite

import (
	"context"
	"io"

	"github.com/walteh/retab/pkg/format"
)

type NoopExternalFormatter struct {
}

var _ ExternalFormatter = (*NoopExternalFormatter)(nil)

func NewNoopExternalFormatProvider() format.Provider {
	return ExternalFormatterToProvider(&NoopExternalFormatter{})
}
func (me *NoopExternalFormatter) Format(_ context.Context, input io.Reader) (io.Reader, func() error) {
	return input, func() error {
		return nil
	}
}

func (me *NoopExternalFormatter) Indent() string {
	return "  "
}

func (me *NoopExternalFormatter) Targets() []string {
	return []string{"*"}
}
