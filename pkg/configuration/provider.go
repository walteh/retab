package configuration

import "context"

type Provider interface {
	GetConfigurationForFileType(ctx context.Context, filename string) (Configuration, error)
}

type Configuration interface {
	UseTabs() bool
	IndentSize() int
	TrimMultipleEmptyLines() bool
}

type basicConfigurationProvider struct {
	tabs                   bool
	indentSize             int
	trimMultipleEmptyLines bool
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

func NewBasicConfigurationProvider(tabs bool, indentSize int, trimMultipleEmptyLines bool) Configuration {
	return &basicConfigurationProvider{
		tabs:                   tabs,
		indentSize:             indentSize,
		trimMultipleEmptyLines: trimMultipleEmptyLines,
	}
}
