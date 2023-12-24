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

func NewTerraformFormatter(cmds ...string) format.Provider {
	cmds = append(cmds, "fmt", "-write=false", "-list=false")

	return NewExecFormatter(&BasicExternalFormatterOpts{
		Indent:  "  ",
		Targets: []string{"*.tf", "*.tfvars", "*.hcl"},
	}, cmds...)
}
