package editorconfig

import (
	"context"
	"strconv"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

type EditorConfigConfiguration struct {
	Definition       *editorconfig.Definition
	parsedIndentSize int
}

type EditorConfigConfigurationProvider struct {
	definitions *editorconfig.Editorconfig
}

type EditorConfigConfigurationDefaults struct {
	Defaults format.Configuration
}

func (me *EditorConfigConfigurationDefaults) GetConfigurationForFileType(ctx context.Context, str string) (format.Configuration, error) {
	return me.Defaults, nil
}

func (me *EditorConfigConfigurationProvider) GetConfigurationForFileType(ctx context.Context, str string) (format.Configuration, error) {
	def, err := me.definitions.GetDefinitionForFilename(str)
	if err != nil {

		return nil, err
	}

	id, err := strconv.Atoi(def.IndentSize)
	if err != nil {
		return nil, errors.Errorf("failed to parse indent size: %w", err)
	}

	return &EditorConfigConfiguration{
		Definition:       def,
		parsedIndentSize: int(id),
	}, nil
}

func NewEditorConfigConfigurationProvider(ctx context.Context, fls afero.Fs) (format.ConfigurationProvider, error) {

	fle, err := fls.Open(".editorconfig")
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("failed to open .editorconfig -- using defaults")
		return &EditorConfigConfigurationDefaults{
			Defaults: format.NewBasicConfigurationProvider(true, 4, true, false),
		}, nil
	}

	x, err2, err := editorconfig.ParseGraceful(fle)
	if err != nil {
		return nil, errors.Errorf("failed to get editorconfig definition: %w", err)
	}
	if err2 != nil {
		return nil, errors.Errorf("failed to parse editorconfig: %w", err2)
	}

	return &EditorConfigConfigurationProvider{
		definitions: x,
	}, nil

}

var _ format.Configuration = &EditorConfigConfiguration{}

func (x *EditorConfigConfiguration) IndentSize() int {
	return x.parsedIndentSize
}

func (x *EditorConfigConfiguration) UseTabs() bool {
	return !strings.Contains(x.Definition.IndentStyle, "space")
}

func (x *EditorConfigConfiguration) TrimMultipleEmptyLines() bool {
	return x.Definition.Raw["trim_multiple_empty_lines"] == "true"
}

func (x *EditorConfigConfiguration) OneBracketPerLine() bool {
	return x.Definition.Raw["one_bracket_per_line"] == "true"
}
