package fmt

import (
	"context"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/dockerfmt"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"github.com/walteh/retab/v2/pkg/format/shfmt"
	"github.com/walteh/retab/v2/pkg/format/yamlfmt"
	"gitlab.com/tozd/go/errors"
)

func getFormatter(ctx context.Context, formatter string, filename string) (format.Provider, error) {

	if formatter == "auto" {
		var autoDetectFormatterGlobs = map[string]string{
			"*.{hcl,hcl2,terraform,tf,tfvars}": "hcl",
			"*.{proto,proto3}":                 "proto",
			"*.{dart}":                         "dart",
			"*.{swift}":                        "swift",
			"*.{yaml,yml}":                     "yaml",
			"*.{sh,bash,zsh,ksh,shell}":        "sh",
			"{Dockerfile,Dockerfile.*}":        "dockerfile",
		}

		for glob, fmt := range autoDetectFormatterGlobs {
			matches, err := doublestar.PathMatch(glob, filename)
			if err != nil {
				return nil, errors.Errorf("globbing: %w", err)
			}
			if matches {
				formatter = fmt
				break
			}
		}
	}

	if formatter == "auto" {
		return nil, errors.New("no formatter found for file path: " + filename)
	}

	switch formatter {
	case "hcl", "hcl2", "terraform", "tf":
		return hclfmt.NewFormatter(), nil
	case "proto", "protobuf":
		return protofmt.NewFormatter(), nil
	case "dart":
		return cmdfmt.NewDartFormatter("dart"), nil
	case "tf-cmd":
		return cmdfmt.NewTerraformFormatter("terraform"), nil
	case "swift":
		return cmdfmt.NewSwiftFormatter("swift"), nil
	case "yaml", "yml":
		return yamlfmt.NewFormatter(), nil
	case "sh", "bash", "zsh", "ksh", "shell":
		return shfmt.NewFormatter(), nil
	case "dockerfile", "docker":
		return dockerfmt.NewFormatter(), nil
	default:
		return nil, errors.New("unknown formatter name: " + formatter)
	}
}
