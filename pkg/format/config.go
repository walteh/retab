package format

import (
	"context"
	"fmt"
)

type ConfigurationProvider interface {
	GetConfigurationForFileType(ctx context.Context, filename string) (Configuration, error)
}

//go:mock
type Configuration interface {
	UseTabs() bool
	IndentSize() int

	Raw() map[string]string
}

func (x *basicConfigurationProvider) GetConfigurationForFileType(ctx context.Context, filename string) (Configuration, error) {
	return x, nil
}

type basicConfigurationProvider struct {
	tabs       bool
	indentSize int
}

func (x *basicConfigurationProvider) UseTabs() bool {
	return x.tabs
}

func (x *basicConfigurationProvider) IndentSize() int {
	return x.indentSize
}

// func (x *basicConfigurationProvider) TrimMultipleEmptyLines() bool {
// 	return x.trimMultipleEmptyLines
// }

// func (x *basicConfigurationProvider) OneBracketPerLine() bool {
// 	return x.oneBracketPerLine
// }

func (x *basicConfigurationProvider) Raw() map[string]string {

	raw := make(map[string]string)
	if x.tabs {
		raw["indent_style"] = "tab"
	} else {
		raw["indent_style"] = "space"
	}
	raw["indent_size"] = fmt.Sprintf("%d", x.indentSize)
	// raw["trim_multiple_empty_lines"] = strconv.FormatBool(x.trimMultipleEmptyLines)
	// raw["one_bracket_per_line"] = strconv.FormatBool(x.oneBracketPerLine)
	return raw
}

func NewBasicConfigurationProvider(tabs bool, indentSize int) Configuration {
	return &basicConfigurationProvider{
		tabs:       tabs,
		indentSize: indentSize,
	}
}

func NewDefaultConfigurationProvider() ConfigurationProvider {
	return &basicConfigurationProvider{
		tabs:       true,
		indentSize: 4,
	}
}
