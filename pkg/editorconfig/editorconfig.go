package editorconfig

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
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

// ConfigOptions represents options for editorconfig resolution
type ConfigOptions struct {
	RawContent string // Raw editorconfig content, if provided will be used directly
	TargetFile string // Path of the file being formatted
}

func NewRawConfigurationProvider(ctx context.Context, rawContent string) (*EditorConfigConfigurationProvider, error) {
	if rawContent == "" {
		return nil, errors.New("raw content is required")
	}

	x, err := editorconfig.Parse(strings.NewReader(rawContent))
	if err != nil {
		return nil, errors.Errorf("getting editorconfig definition: %w", err)
	}
	return &EditorConfigConfigurationProvider{definitions: x}, nil

}

func NewDynamicConfigurationProvider(ctx context.Context, rawContent string) (*EditorConfigConfigurationProvider, error) {
	if rawContent != "" {
		// If raw content is provided, parse it directly
		x, err := editorconfig.Parse(strings.NewReader(rawContent))
		if err != nil {
			return nil, errors.Errorf("getting editorconfig definition: %w", err)
		}
		return &EditorConfigConfigurationProvider{definitions: x}, nil
	}

	return &EditorConfigConfigurationProvider{}, nil
}

func (me *EditorConfigConfigurationProvider) GetConfigurationForFileType(ctx context.Context, targetFile string) (format.Configuration, error) {
	var def *editorconfig.Definition
	var err error

	if me.definitions != nil {
		def, err = me.definitions.GetDefinitionForFilename(filepath.Base(targetFile))
		if err != nil {
			return nil, errors.Errorf("getting editorconfig definition: %w", err)
		}
	} else {
		// Otherwise, let the library handle auto-resolution
		def, err = editorconfig.GetDefinitionForFilenameWithConfigname(targetFile, ".editorconfig")
		if err != nil {
			return nil, errors.Errorf("getting editorconfig definition: %w", err)
		}
	}

	if def.IndentSize == "" {
		def.IndentSize = "4"
	}

	id, err := strconv.Atoi(def.IndentSize)
	if err != nil {
		return nil, errors.Errorf("parsing indent size: %w", err)
	}

	return &EditorConfigConfiguration{
		Definition:       def,
		parsedIndentSize: int(id),
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

func (x *EditorConfigConfiguration) Raw() map[string]string {
	raw := make(map[string]string)
	for k, v := range x.Definition.Raw {
		raw[k] = v
	}
	return raw
}
