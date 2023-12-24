package editorconfig

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/walteh/retab/pkg/configuration"
)

type EditorConfigConfigurationProvider struct {
	Definition       *editorconfig.Definition
	parsedIndentSize int
}

func NewEditorConfigConfigurationProvider(_ context.Context, filename string) (configuration.Provider, error) {
	x, err := editorconfig.GetDefinitionForFilename(filename)
	if err != nil {
		dir := filepath.Dir(filename)
		x, err = editorconfig.GetDefinitionForFilename(filepath.Join(dir, ".editorconfig"))
		if err != nil {
			return nil, err
		}
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

func (x *EditorConfigConfigurationProvider) TrimMultipleEmptyLines() bool {
	return x.Definition.Raw["trim_multiple_empty_lines"] == "true"
}
