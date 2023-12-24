package editorconfig

import (
	"context"
	"strconv"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/terrors"
)

type EditorConfigConfiguration struct {
	Definition       *editorconfig.Definition
	parsedIndentSize int
}

type EditorConfigConfigurationProvider struct {
	definitions *editorconfig.Editorconfig
}

type EditorConfigConfigurationDefaults struct {
	Defaults configuration.Configuration
}

func (me *EditorConfigConfigurationDefaults) GetConfigurationForFileType(ctx context.Context, str string) (configuration.Configuration, error) {
	return me.Defaults, nil
}

func (me *EditorConfigConfigurationProvider) GetConfigurationForFileType(ctx context.Context, str string) (configuration.Configuration, error) {
	def, err := me.definitions.GetDefinitionForFilename(str)
	if err != nil {

		return nil, err
	}

	id, err := strconv.Atoi(def.IndentSize)
	if err != nil {
		return nil, terrors.Wrap(err, "failed to parse indent size")
	}

	return &EditorConfigConfiguration{
		Definition:       def,
		parsedIndentSize: int(id),
	}, nil
}

func NewEditorConfigConfigurationProvider(ctx context.Context, fls afero.Fs) (configuration.Provider, error) {

	fle, err := fls.Open(".editorconfig")
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("failed to open .editorconfig -- using defaults")
		return &EditorConfigConfigurationDefaults{
			Defaults: configuration.NewBasicConfigurationProvider(true, 4, true),
		}, nil
	}

	x, err2, err := editorconfig.ParseGraceful(fle)
	if err != nil {
		return nil, terrors.Wrap(err, "failed to get editorconfig definition")
	}
	if err2 != nil {
		return nil, terrors.Wrap(err2, "failed to parse editorconfig")
	}

	return &EditorConfigConfigurationProvider{
		definitions: x,
	}, nil

}

var _ configuration.Configuration = &EditorConfigConfiguration{}

func (x *EditorConfigConfiguration) IndentSize() int {
	return x.parsedIndentSize
}

func (x *EditorConfigConfiguration) UseTabs() bool {
	return !strings.Contains(x.Definition.IndentStyle, "space")
}

func (x *EditorConfigConfiguration) TrimMultipleEmptyLines() bool {
	return x.Definition.Raw["trim_multiple_empty_lines"] == "true"
}
