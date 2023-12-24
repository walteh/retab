package protowrite

import (
	"context"
	"io"

	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.proto", "*.proto3"}
}

func (me *Formatter) Format(_ context.Context, cfg configuration.Configuration, read io.Reader) (io.Reader, error) {

	fileNode, err := parser.Parse("retab.protobuf-parser", read, reporter.NewHandler(nil))
	if err != nil {
		return nil, err
	}

	read, write := io.Pipe()

	fmtr := newFormatter(write, fileNode, cfg)

	go func() {
		if err := fmtr.Run(); err != nil {
			err := write.CloseWithError(err)
			if err != nil {
				panic(err)
			}
			return
		}
		if err := write.Close(); err != nil {
			panic(err)
		}
	}()

	return read, nil
}
