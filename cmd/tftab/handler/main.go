package handler

// `hclFmt` commmand recursively looks for hcl files in the directory tree starting at workingDir, and formats them
// based on the language style guides provided by Hashicorp. This is done using the official hcl2 library.

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-zglob"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/tftab/pkg/configuration/editorconfig"
	"github.com/walteh/tftab/pkg/hclwrite"
)

type Handler struct {
	File       string `arg:"" default:"" name:"file" help:"The hcl file to format."`
	WorkingDir string `flag:"" name:"working-dir" default:"." help:"The directory to recursively search for hcl files."`
}

func (me *Handler) Run(ctx context.Context, fs afero.Fs) error {

	// handle when option specifies a particular file
	if me.File != "" {
		cfg, err := editorconfig.NewEditorConfigConfigurationProvider(ctx, me.File)
		if err != nil {
			return err
		}

		if !filepath.IsAbs(me.File) {
			me.File = filepath.Join(me.WorkingDir, me.File)
		}
		zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl file at: %s.", me.File)
		return hclwrite.Process(ctx, cfg, fs, me.File)
	}

	zerolog.Ctx(ctx).Debug().Msgf("Formatting hcl files from the directory tree %s.", me.WorkingDir)

	// zglob normalizes paths to "/"
	extensions := []string{"*.hcl", "*.tf", "*.tfvars", "*.hcl2"}
	var files []string

	for _, ext := range extensions {
		pattern := filepath.Join(me.WorkingDir, "**", ext)
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
		err = hclwrite.Process(ctx, cfg, fs, filename)
		if err != nil {
			formatErrors = multierror.Append(formatErrors, err)
		}
	}

	return formatErrors.ErrorOrNil()
}
