package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/editorconfig"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"gitlab.com/tozd/go/errors"
)

type Handler struct {
	filename            string
	formatter           string // auto, hcl, proto, dart, tf
	ToStdout            bool
	FromStdin           bool
	editorconfigContent string
}

func NewFmtCommand() *cobra.Command {
	me := &Handler{}

	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format files with the hcl golang library, but with tabs",
	}

	cmd.Flags().StringVar(&me.formatter, "formatter", "auto", "the formatter to use")
	cmd.Flags().BoolVar(&me.ToStdout, "stdout", false, "write to stdout instead of file")
	cmd.Flags().BoolVar(&me.FromStdin, "stdin", false, "read from stdin instead of file")

	cmd.Flags().StringVar(&me.editorconfigContent, "editorconfig-content", "", "editorconfig content (optional)")
	cmd.Args = cobra.ExactArgs(1)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		me.filename = args[0]
		return me.Run(cmd.Context())
	}

	return cmd
}

func (me *Handler) getFormatter(ctx context.Context) (format.Provider, error) {
	if me.formatter == "auto" {
		formatters := []format.Provider{
			hclfmt.NewFormatter(),
			protofmt.NewFormatter(),
			cmdfmt.NewDartFormatter("dart"),
			cmdfmt.NewTerraformFormatter("terraform"),
			cmdfmt.NewSwiftFormatter("swift"),
		}
		fmtr, err := format.AutoDetectFormatter(me.filename, formatters)
		if err != nil {
			return nil, errors.Errorf("auto-detecting formatter: %w", err)
		}
		if fmtr == nil {
			return nil, errors.Errorf("no formatters found for file '%s'", me.filename)
		}
		return fmtr, nil
	}

	switch me.formatter {
	case "hcl":
		return hclfmt.NewFormatter(), nil
	case "proto":
		return protofmt.NewFormatter(), nil
	case "dart":
		return cmdfmt.NewDartFormatter("dart"), nil
	case "tf":
		return cmdfmt.NewTerraformFormatter("terraform"), nil
	case "swift":
		return cmdfmt.NewSwiftFormatter("swift"), nil
	default:
		return nil, errors.New("invalid formatter")
	}
}

func (me *Handler) Run(ctx context.Context) error {
	fs := afero.NewOsFs()

	// Setup editorconfig with either raw content or auto-resolution
	cfgProvider, err := editorconfig.NewDynamicConfigurationProvider(ctx, me.editorconfigContent)
	if err != nil {
		return errors.Errorf("creating configuration provider: %w", err)
	}

	fmtr, err := me.getFormatter(ctx)
	if err != nil {
		return err
	}

	var input io.Reader
	if me.FromStdin {
		input = os.Stdin
	} else {
		file, err := fs.Open(me.filename)
		if err != nil {
			return errors.Errorf("opening file: %w", err)
		}
		defer file.Close()
		input = file
	}

	r, err := format.Format(ctx, fmtr, cfgProvider, me.filename, input)
	if err != nil {
		return errors.Errorf("formatting content: %w", err)
	}

	if me.ToStdout || me.FromStdin {
		_, err = io.Copy(os.Stdout, r)
		return err
	}

	err = afero.WriteReader(fs, me.filename, r)
	if err != nil {
		return errors.Errorf("writing formatted file: %w", err)
	}

	return nil
}
