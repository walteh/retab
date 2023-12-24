package hclwrite

import (
	"context"
	"io"

	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.hcl", "*.tf", "*.tfvars", "*.hcl2"}
}

func (me *Formatter) Format(ctx context.Context, cfg configuration.Configuration, read io.Reader) (io.Reader, error) {

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
