package fmt

// `hclFmt` commmand recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-zglob"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/tftab/pkg/cli"
	"github.com/walteh/tftab/pkg/configuration/editorconfig"
	"github.com/walteh/tftab/pkg/hclwrite"
)

var _ cli.Cobraface = (*Handler)(nil)

type Handler struct {
	File       string `arg:"" default:"" name:"file" help:"The hcl file to format."`
	WorkingDir string `name:"working-dir" help:"The working directory to use. Defaults to the current directory."`
}

func (me *Handler) Define(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt",
		Short: "format hcl files with the official hcl2 library, but with tabs",
	}
	cmd.Args = cobra.ExactArgs(1)

	cmd.Flags().StringVar(&me.WorkingDir, "working-dir", "", "The working directory to use. Defaults to the current directory.")

	return cmd
}

func (me *Handler) Inject(ctx context.Context, cmd *cobra.Command, file []string) error {

	me.File = file[0]

	return nil

}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {

	isDir, err := afero.IsDir(fs, me.File)
	if err != nil {
		return err
	}

	zerolog.Ctx(ctx).Trace().Any("handler", me).Msg("Running fmt command.")

	// handle when option specifies a particular file
	if !isDir {
		cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, me.File)
		if err != nil {
			return err
		}

		if !filepath.IsAbs(me.File) {
			me.File = filepath.Join(me.WorkingDir, me.File)
		}
		zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl file at: %s.", me.File)
		return hclwrite.Format(ctx, cfg, fs, me.File)
	}

	zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl files from the directory tree %s %s", me.WorkingDir, me.File)

	// zglob normalizes paths to "/"
	extensions := []string{"*.hcl", "*.tf", "*.tfvars", "*.hcl2"}
	var files []string

	for _, ext := range extensions {
		pattern := filepath.Join(me.WorkingDir, me.File, "**", ext)
		matches, err := zglob.Glob(pattern)
		if err != nil {
			return err
		}
		files = append(files, matches...)
	}

	var formatErrors *multierror.Error
	for _, filename := range files {
		cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, filename)
		if err != nil {
			return err
		}
		err = hclwrite.Format(ctx, cfg, fs, filename)
		if err != nil {
			formatErrors = multierror.Append(formatErrors, err)
		}
	}

	return formatErrors.ErrorOrNil()
}
