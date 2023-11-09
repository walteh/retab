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

func NewDartFileFormatter(file string, cmds ...string) format.Provider {
	cmds = append(cmds, "format", "--output", "show", "--summary", "none", "--fix", file)

	return NewExecFormatter(&BasicExternalFormatterOpts{
		Indent:  "  ",
		Targets: []string{"*.dart"},
	}, cmds...)
}
