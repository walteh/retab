package externalwrite

import (
	"github.com/walteh/retab/pkg/format"
)

func NewDartFormatter(cmds ...string) format.Provider {
	cmds = append(cmds, "format", "--output", "show", "--summary", "none", "--fix")

	return NewExecFormatter(&BasicExternalFormatterOpts{
		Indent:  "  ",
		Targets: []string{"*.dart"},
	}, cmds...)
}
