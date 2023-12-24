package configuration

import "context"

type Provider interface {
	GetConfigurationForFileType(ctx context.Context, filename string) (Configuration, error)
}

type Configuration interface {
	UseTabs() bool
	IndentSize() int
	TrimMultipleEmptyLines() bool
	OneBracketPerLine() bool
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
