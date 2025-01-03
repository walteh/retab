package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
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
	filename  string
	formatter string // auto, hcl, proto, dart, tf
	ToStdout  bool
	FromStdin bool
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
	// the glob will will be argument one
	cmd.Args = cobra.ExactArgs(1)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		me.filename = args[0]
		return me.Run(cmd.Context())
	}

	return cmd
}

func (me *Handler) Run(ctx context.Context) error {
	var fmtr format.Provider

	fs := afero.NewOsFs()

	cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, fs)
	if err != nil {
		return errors.Errorf("failed to create editorconfig configuration provider: %w", err)
	}

	if me.formatter == "auto" {
		pfmters := []format.Provider{}
		pfmters = append(pfmters, hclfmt.NewFormatter())
		pfmters = append(pfmters, protofmt.NewFormatter())
		pfmters = append(pfmters, cmdfmt.NewDartFormatter("dart"))
		pfmters = append(pfmters, cmdfmt.NewTerraformFormatter("terraform"))
		for _, pfmtr := range pfmters {
			for _, target := range pfmtr.Targets() {
				basename := filepath.Base(me.filename)
				ok, err := doublestar.Match(target, basename)
				if err != nil {
					return errors.Errorf("failed to match glob: %w", err)
				}
				if ok {
					fmtr = pfmtr
					break
				}
			}
		}
	} else if me.formatter == "hcl" {
		fmtr = hclfmt.NewFormatter()
	} else if me.formatter == "proto" {
		fmtr = protofmt.NewFormatter()
	} else if me.formatter == "dart" {
		fmtr = cmdfmt.NewDartFormatter("dart")
	} else if me.formatter == "tf" {
		fmtr = cmdfmt.NewTerraformFormatter("terraform")
	} else {
		return errors.New("invalid formatter")
	}

	if fmtr == nil {
		return errors.Errorf("no formatters found for file '%s'", me.filename)
	}

	var input io.Reader
	if me.FromStdin {
		input = os.Stdin
	} else {
		file, err := fs.Open(me.filename)
		if err != nil {
			return errors.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		input = file
	}

	r, err := format.Format(ctx, fmtr, cfg, me.filename, input)
	if err != nil {
		return errors.Errorf("failed to format file: %w", err)
	}

	if me.ToStdout || me.FromStdin {
		io.Copy(os.Stdout, r)
		return nil
	}

	err = afero.WriteReader(fs, me.filename, r)
	if err != nil {
		return errors.Errorf("failed to write formatted file: %w", err)
	}

	return nil
}
