package hclwrite

import (
	"context"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/tftab/pkg/configuration"
)

func Format(cfg configuration.Provider, src []byte) (io.Reader, error) {
	tokens := lexConfig(src)
	tokens.format()
	r, w := io.Pipe()
	go func() {
		_, err := tokens.WriteTo(w, cfg)
		if err != nil {
			w.CloseWithError(err)
			return
		} else {
			w.Close()
		}
	}()
	return r, nil
}

// Process uses the hcl2 library to format the hcl file. This will attempt to parse the HCL file first to
// ensure that there are no syntax errors, before attempting to format it.
func Process(ctx context.Context, cfg configuration.Provider, fs afero.Fs, fle string) error {
	zerolog.Ctx(ctx).Debug().Msgf("Formatting %s", fle)

	contents, err := afero.ReadFile(fs, fle)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("Error reading %s", fle)
		return err
	}

	err = checkErrors(ctx, contents, fle)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("Error parsing %s", fle)
		return err
	}

	newContents, err := Format(cfg, contents)
	if err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msgf("Error formatting %s", fle)
		return err
	}

	zerolog.Ctx(ctx).Info().Msgf("%s was updated", fle)

	return afero.WriteReader(fs, fle, newContents)
}

// checkErrors takes in the contents of a hcl file and looks for syntax errors.
func checkErrors(ctx context.Context, contents []byte, fle string) error {
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL(contents, fle)
	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 0, true)
	defer func() {
		err := diagWriter.WriteDiagnostics(diags)
		if err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Msgf("Error writing diagnostics for %s", fle)
		}
	}()
	if diags.HasErrors() {
		return diags
	}
	return nil
}
