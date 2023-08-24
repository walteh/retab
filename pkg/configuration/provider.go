package configuration

type Provider interface {
	UseTabs() bool
	IndentSize() int
}
