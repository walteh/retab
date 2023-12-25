package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"strings"

	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/externalwrite"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/retab/pkg/protowrite"
	"github.com/walteh/snake"
)

func Runner() snake.Runner {
	return snake.GenRunCommand_In05_Out01(&Handler{})
}

type Handler struct {
	CustomCommand      string   `default:"" usage:"use custom formating command"`
	TargetRestrictions []string `usage:"file patterns to match - for example, *.tfvars"`
	CustomCommandIdent int      `default:"2" usage:"the number of spaces the custom formatter uses for indentation - for example, for terraform format you would put 2"`
}

func (me *Handler) Name() string {
	return "fmt"
}

func (me *Handler) Description() string {
	return "format files with the hcl golang library, but with tabs"
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, fle afero.File, ecfg configuration.Provider, out snake.Stdout) error {
	fmtrs := []format.Provider{}

	if me.CustomCommand != "" {
		indentstr := ""
		for i := 0; i < me.CustomCommandIdent; i++ {
			indentstr += " "
		}
		fmtrs = append(fmtrs, externalwrite.NewExecFormatter(&externalwrite.BasicExternalFormatterOpts{
			Indent:  indentstr,
			Targets: me.TargetRestrictions,
		}, strings.Split(me.CustomCommand, " ")...))
	} else {
		fmtrs = append(fmtrs, hclwrite.NewFormatter())
		fmtrs = append(fmtrs, protowrite.NewFormatter())
		fmtrs = append(fmtrs, externalwrite.NewDartFormatter("dart"))
		fmtrs = append(fmtrs, externalwrite.NewTerraformFormatter("terraform"))
	}

	for _, fmtr := range fmtrs {
		err := format.Format(ctx, fmtr, ecfg, fs, fle)
		if err != nil {
			return err
		}
	}

	return nil
}
