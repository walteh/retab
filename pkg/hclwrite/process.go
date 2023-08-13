package hclwrite

import (
	"context"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func Format(src []byte) (io.Reader, error) {
	tokens := lexConfig(src)
	tokens.format()
	r, w := io.Pipe()
	go func() {
		_, err := tokens.WriteTo(w)
		if err != nil {
			w.CloseWithError(err)
			return
		} else {
			w.Close()
		}
	}()
	return r, nil
}

// formatTgHCL uses the hcl2 library to format the hcl file. This will attempt to parse the HCL file first to
// ensure that there are no syntax errors, before attempting to format it.
func Process(ctx context.Context, fs afero.Fs, tgHclFile string) error {
	zerolog.Ctx(ctx).Debug().Msgf("Formatting %s", tgHclFile)

	// info, err := fs.Stat(tgHclFile)
	// if err != nil {
	// 	zerolog.Ctx(ctx).Error().Err(err).Msgf("Error retrieving file info of %s", tgHclFile)
	// 	return err
	// }

	contents, err := afero.ReadFile(fs, tgHclFile)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("Error reading %s", tgHclFile)
		return err
	}

	err = checkErrors(ctx, contents, tgHclFile)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("Error parsing %s", tgHclFile)
		return err
	}

	newContents, err := Format(contents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("Error formatting %s", tgHclFile)
		return err
	}

	// fileUpdated := !bytes.Equal(newContents, contents)

	// if fileUpdated {

	// }
	zerolog.Ctx(ctx).Info().Msgf("%s was updated", tgHclFile)

	return afero.WriteReader(fs, tgHclFile, newContents)
}

// checkErrors takes in the contents of a hcl file and looks for syntax errors.
func checkErrors(ctx context.Context, contents []byte, tgHclFile string) error {
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL(contents, tgHclFile)
	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)
	defer func() {
		err := diagWriter.WriteDiagnostics(diags)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msgf("Error writing diagnostics for %s", tgHclFile)
		}
	}()
	if diags.HasErrors() {
		return diags
	}
	return nil
}
