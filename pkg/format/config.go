package format

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"text/tabwriter"
)

type ConfigurationProvider interface {
	GetConfigurationForFileType(ctx context.Context, filename string) (Configuration, error)
}

//go:mock
type Configuration interface {
	UseTabs() bool
	IndentSize() int
	TrimMultipleEmptyLines() bool
	OneBracketPerLine() bool
	Raw() map[string]string
}

func BuildTabWriter(cfg Configuration, writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, cfg.IndentSize(), 1, ' ', tabwriter.TabIndent|tabwriter.StripEscape|tabwriter.DiscardEmptyColumns)
}

type basicConfigurationProvider struct {
	tabs                   bool
	indentSize             int
	trimMultipleEmptyLines bool
	oneBracketPerLine      bool
}

func (x *basicConfigurationProvider) UseTabs() bool {
	return x.tabs
}

func (x *basicConfigurationProvider) IndentSize() int {

	return x.indentSize
}

func (x *basicConfigurationProvider) TrimMultipleEmptyLines() bool {
	return x.trimMultipleEmptyLines
}

func (x *basicConfigurationProvider) OneBracketPerLine() bool {
	return x.oneBracketPerLine
}

func (x *basicConfigurationProvider) Raw() map[string]string {

	raw := make(map[string]string)
	if x.tabs {
		raw["indent_style"] = "tab"
	} else {
		raw["indent_style"] = "space"
	}
	raw["indent_size"] = fmt.Sprintf("%d", x.indentSize)
	raw["trim_multiple_empty_lines"] = strconv.FormatBool(x.trimMultipleEmptyLines)
	raw["one_bracket_per_line"] = strconv.FormatBool(x.oneBracketPerLine)
	return raw
}

func NewBasicConfigurationProvider(tabs bool, indentSize int, trimMultipleEmptyLines bool, onebracket bool) Configuration {
	return &basicConfigurationProvider{
		tabs:                   tabs,
		indentSize:             indentSize,
		trimMultipleEmptyLines: trimMultipleEmptyLines,
		oneBracketPerLine:      onebracket,
	}
}
