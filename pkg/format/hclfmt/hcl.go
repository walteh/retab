package hclfmt

import (
	"context"
	"io"

	"github.com/walteh/retab/v2/pkg/format"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.hcl", "*.hcl2", "*.tf", "*.tfvars"}
}

func (me *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {

	reads, err := io.ReadAll(read)
	if err != nil {
		return nil, err
	}

	err = checkErrors(ctx, reads, "")
	if err != nil {
		return nil, err
	}

	newContents, err := FormatBytes(cfg, reads)
	if err != nil {
		return nil, err
	}

	return newContents, nil
}
