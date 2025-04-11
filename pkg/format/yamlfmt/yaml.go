package yamlfmt

import (
	"bytes"
	"context"
	"io"
	"strconv"

	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/formatters/basic"
	"github.com/mitchellh/mapstructure"
	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Targets() []string {
	return []string{"*.yaml", "*.yml"}
}

func (me *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {

	reads, err := io.ReadAll(read)
	if err != nil {
		return nil, err
	}

	formatter, err := getFormatter(cfg)
	if err != nil {
		return nil, err
	}

	// engine, err := getEngine(formatter)
	// if err != nil {
	// 	return nil, err
	// }

	out, err := formatter.Format(reads)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(out), nil

}

func getFormatter(cfg format.Configuration) (*basic.BasicFormatter, error) {
	def := basic.DefaultConfig()

	if raw := cfg.Raw(); raw != nil {
		if maxLineLength, ok := raw["max_line_length"]; ok {
			if val, err := strconv.Atoi(maxLineLength); err == nil {
				def.LineLength = val
				delete(raw, "max_line_length")
			}
		}
		if padComments, ok := raw["pad_line_comments"]; ok {
			if val, err := strconv.Atoi(padComments); err == nil {
				def.PadLineComments = val
				delete(raw, "pad_line_comments")
			}
		}
		err := mapstructure.Decode(raw, &def)
		if err != nil {
			return nil, errors.Errorf("decoding raw editor config: %w", err)
		}
	}

	// Convert editor config settings to YAML formatter settings
	def.Indent = cfg.IndentSize()
	def.TrimTrailingWhitespace = true         // Always trim trailing whitespace
	def.EOFNewline = true                     // Always ensure newline at EOF
	def.LineEnding = yamlfmt.LineBreakStyleLF // Always use LF line breaks
	def.ScanFoldedAsLiteral = false
	def.IndentlessArrays = false
	def.IndentRootArray = true

	// Handle indentation style
	if cfg.UseTabs() {
		def.Indent = 4 * cfg.IndentSize() // When using tabs, we still need to set an indent size for internal spacing
	}

	def.ArrayIndent = cfg.IndentSize()

	// Handle line breaks
	def.RetainLineBreaks = false
	def.RetainLineBreaksSingle = true

	f := basic.BasicFormatter{
		Config:       def,
		YAMLFeatures: basic.ConfigureYAMLFeaturesFromConfig(def),
		Features:     basic.ConfigureFeaturesFromConfig(def),
	}

	return &f, nil
}
