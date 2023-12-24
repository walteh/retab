package wrap

import (
	"context"
	"strings"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In04_Out01(&Handler{})
}

type Handler struct {
	Dart                     bool     `default:"false" usage:"use dart formating presets"`
	Terraform                bool     `default:"false" usage:"use terraform formating presets"`
	Command                  string   `default:"" usage:"use custom formating command"`
	CommandTargets           []string `default:"*" usage:"file patterns to match - for example, *.dart"`
	CommandUndesirableIndent int      `default:"2" usage:"the number of spaces the custom formatter uses for indentation - for example, for terraform format you would put 2"`
}

func (me *Handler) Name() string {
	return "wrap"
}

func (me *Handler) Description() string {
	return "wraps an external formatter you have installed on your system"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider) error {
	var fmtr format.Provider

	if me.Dart || me.Terraform {
		cmd := ""
		if me.Dart {
			if me.Command != "" {
				cmd = "dart"
			}
			fmtr = externalwrite.NewDartFormatter(strings.Split(cmd, " ")...)
		} else {
			if me.Command != "" {
				cmd = "terraform"
			}
			fmtr = externalwrite.NewTerraformFormatter(strings.Split(cmd, " ")...)
		}
	} else {
		identstr := ""
		for i := 0; i < me.CommandUndesirableIndent; i++ {
			identstr += " "
		}

		fmtr = externalwrite.NewExecFormatter(&externalwrite.BasicExternalFormatterOpts{
			Indent:  identstr,
			Targets: me.CommandTargets,
		}, strings.Split(me.Command, " ")...)

	}

	return format.Format(ctx, fmtr, ecfg, fs, fle)

}
