package format

import (
	"context"
	"io"
	"text/tabwriter"
)

type ConfigurationProvider interface {
	GetConfigurationForFileType(ctx context.Context, filename string) (Configuration, error)
}

type Configuration interface {
	UseTabs() bool
	IndentSize() int
	TrimMultipleEmptyLines() bool
	OneBracketPerLine() bool
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

func NewBasicConfigurationProvider(tabs bool, indentSize int, trimMultipleEmptyLines bool, onebracket bool) Configuration {
	return &basicConfigurationProvider{
		tabs:                   tabs,
		indentSize:             indentSize,
		trimMultipleEmptyLines: trimMultipleEmptyLines,
		oneBracketPerLine:      onebracket,
	}
}
