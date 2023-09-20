// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type ExperimentalFeatures struct {
	ValidateOnSave        bool `mapstructure:"validateOnSave"`
	PrefillRequiredFields bool `mapstructure:"prefillRequiredFields"`
}

type Indexing struct {
	IgnoreDirectoryNames []string `mapstructure:"ignoreDirectoryNames"`
	IgnorePaths          []string `mapstructure:"ignorePaths"`
}

type Options struct {
	CommandPrefix string   `mapstructure:"commandPrefix"`
	Indexing      Indexing `mapstructure:"indexing"`

	Path string `mapstructure:"path"`

	// ExperimentalFeatures encapsulates experimental features users can opt into.
	ExperimentalFeatures ExperimentalFeatures `mapstructure:"experimentalFeatures"`

	IgnoreSingleFileWarning bool `mapstructure:"ignoreSingleFileWarning"`
}

func (o *Options) Validate() error {
	if o.Path != "" {
		path := o.Path
		if !filepath.IsAbs(path) {
			return fmt.Errorf("Expected absolute path for binary, got %q", path)
		}
		stat, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("Unable to find binary: %s", err)
		}
		if stat.IsDir() {
			return fmt.Errorf("Expected a binary, got a directory: %q", path)
		}
	}

	if len(o.Indexing.IgnoreDirectoryNames) > 0 {
		for _, directory := range o.Indexing.IgnoreDirectoryNames {
			if strings.Contains(directory, string(filepath.Separator)) {
				return fmt.Errorf("expected directory name, got a path: %q", directory)
			}
		}
	}

	return nil
}

type DecodedOptions struct {
	Options    *Options
	UnusedKeys []string
}

func DecodeOptions(input interface{}) (*DecodedOptions, error) {
	var md mapstructure.Metadata
	var options Options

	config := &mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &options,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(err)
	}

	if err := decoder.Decode(input); err != nil {
		return nil, err
	}

	return &DecodedOptions{
		Options:    &options,
		UnusedKeys: md.Unused,
	}, nil
}
