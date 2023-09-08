package hclwrite

import (
	"context"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
)

// WriteTo takes an io.Writer and writes the bytes for each token to it,
// along with the spacing that separates each token. In other words, this
// allows serializing the tokens to a file or other such byte stream.
func (ts Tokens) WriteTo(wr io.Writer, cfg configuration.Provider) (int64, error) {
	// We know we're going to be writing a lot of small chunks of repeated
	// space characters, so we'll prepare a buffer of these that we can
	// easily pass to wr.Write without any further allocation.
	spaces := make([]byte, 40)
	for i := range spaces {
		spaces[i] = ' '
	}

	tabs := make([]byte, 40)
	for i := range tabs {
		tabs[i] = '\t'
	}

	var n int64
	var err error
	for _, token := range ts {
		if err != nil {
			return n, err
		}

		// Write the leading tabs, if any
		for tabsBefore := token.TabsBefore; tabsBefore > 0; tabsBefore -= len(tabs) {
			thisChunk := tabsBefore
			if thisChunk > len(tabs) {
				thisChunk = len(tabs)
			}
			var thisN int
			if cfg.UseTabs() {
				thisN, err = wr.Write(tabs[:thisChunk])

			} else {
				thisN, err = wr.Write(spaces[:thisChunk*cfg.IndentSize()])
			}
			n += int64(thisN)
			if err != nil {
				return n, err
			}
		}

		for spacesBefore := token.SpacesBefore; spacesBefore > 0; spacesBefore -= len(spaces) {
			thisChunk := spacesBefore
			if thisChunk > len(spaces) {
				thisChunk = len(spaces)
			}
			var thisN int
			thisN, err = wr.Write(spaces[:thisChunk])
			n += int64(thisN)
			if err != nil {
				return n, err
			}
		}

		var thisN int
		thisN, err = wr.Write(token.Bytes)
		n += int64(thisN)
	}

	return n, err
}

func FormatBytes(cfg configuration.Provider, src []byte) (io.Reader, error) {
	tokens := lexConfig(src)
	tokens.format()
	r, w := io.Pipe()
	go func() {
		_, err := tokens.WriteTo(w, cfg)
		if err != nil {
			w.CloseWithError(err)
			return
		}
		if err := w.Close(); err != nil {
			panic(err)
		}
	}()
	return r, nil
}

// Process uses the hcl2 library to format the hcl file. This will attempt to parse the HCL file first to
// ensure that there are no syntax errors, before attempting to format it.
func Format(ctx context.Context, cfg configuration.Provider, fs afero.Fs, fle string) error {
	zerolog.Ctx(ctx).Debug().Any("config", cfg).Msgf("Formatting %s", fle)

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

	newContents, err := FormatBytes(cfg, contents)
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
