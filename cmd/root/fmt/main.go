package fmt

// `hclFmt` command recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"errors"

	"github.com/mattn/go-zglob"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/pkg/bufwrite"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/retab/pkg/format"
	"github.com/walteh/retab/pkg/hclwrite"
	"github.com/walteh/snake"
)

var _ snake.Cobrad = (*Handler)(nil)

type Handler struct {
	File       string `arg:"" default:" " name:"file" help:"The hcl file to format."`
	WorkingDir string `name:"working-dir" help:"The working directory to use. Defaults to the current directory."`
}

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt <file>",
		Short: "format hcl files with the official hcl2 library, but with tabs",
	}

	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a file argument")
		}
		me.File = args[0]
		if me.File == "" {
			return errors.New("no file provided")
		}
		return nil
	}

	cmd.Flags().StringVar(&me.WorkingDir, "working-dir", "", "The working directory to use. Defaults to the current directory.")

	return cmd
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs, ecfg configuration.Provider) error {

	fmtr := hclwrite.NewHclFormatter()
	bufr := bufwrite.NewBufFormatter()

	for _, target := range bufr.Targets() {
		// targets are glob patterns
		matches, err := zglob.Glob(target)
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			continue
		}

		return format.Format(ctx, bufr, ecfg, fs, me.File, me.WorkingDir)
	}

	for _, target := range fmtr.Targets() {
		// targets are glob patterns
		matches, err := zglob.Glob(target)
		if err != nil {
			return err
		}

		if len(matches) == 0 {
			continue
		}

		return format.Format(ctx, fmtr, ecfg, fs, me.File, me.WorkingDir)
	}

	return errors.New("no targets found")
}
