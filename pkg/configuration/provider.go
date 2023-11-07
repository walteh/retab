package configuration

type Provider interface {
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

func NewBasicConfigurationProvider(tabs bool, indentSize int, trimMultipleEmptyLines bool) Provider {
	return &basicConfigurationProvider{
		tabs:                   tabs,
		indentSize:             indentSize,
		trimMultipleEmptyLines: trimMultipleEmptyLines,
	}
}
