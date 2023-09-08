package editorconfig

import (
	"context"
	"strconv"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/walteh/retab/pkg/configuration"
)

type EditorConfigConfigurationProvider struct {
	Definition       *editorconfig.Definition
	parsedIndentSize int
}

func NewEditorConfigConfigurationProvider(_ context.Context, filename string) (*EditorConfigConfigurationProvider, error) {
	x, err := editorconfig.GetDefinitionForFilename(filename)
	if err != nil {
		return nil, err
	}

	ids, err := strconv.Atoi(x.IndentSize)
	if err != nil {
		return nil, err
	}

	return &EditorConfigConfigurationProvider{
		Definition:       x,
		parsedIndentSize: ids,
	}, nil

}

var _ configuration.Provider = &EditorConfigConfigurationProvider{}

func (x *EditorConfigConfigurationProvider) IndentSize() int {
	return x.parsedIndentSize
}

func (x *EditorConfigConfigurationProvider) UseTabs() bool {
	return !strings.Contains(x.Definition.IndentStyle, "space")
}
