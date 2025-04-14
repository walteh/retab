package fmt

import (
	"context"
	"path/filepath"
	"strings"

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

	if formatter == "auto" || formatter == "" {

		// // Wrap in a buffered reader
		// br := bufio.NewReader(reader)

		// // Peek at the first 250 bytes without advancing the reader
		// peeked, _ := br.Peek(250)

		// langs := enry.GetLanguages(filename, peeked)
		// if len(langs) == 0 {
		// 	return nil, errors.Errorf("failed to get language")
		// }

		// zerolog.Ctx(ctx).Info().Strs("languages", langs).Msg("detected languages")

		// for _, lang := range langs {
		// 	fmtr, ok := getFormatterByLanguage(lang)
		// 	if !ok {
		// 		continue
		// 	}

		// 	zerolog.Ctx(ctx).Info().Msgf("using formatter for language [%s]", lang)
		// 	return fmtr, nil
		// }

		var backupAutoDetectFormatterGlobs = map[string]string{
			"*.{hcl,hcl2,terraform,tf,tfvars}": "hcl",
			"*.{proto,proto3}":                 "proto",
			"*.dart":                           "dart",
			"*.swift":                          "swift",
			"*.{yaml,yml}":                     "yaml",
			"*.{sh,bash,zsh,ksh,shell}":        "sh",
			"{Dockerfile,Dockerfile.*}":        "dockerfile",
		}

		for glob, fmt := range backupAutoDetectFormatterGlobs {
			matches, err := doublestar.PathMatch(glob, filepath.Base(filename))
			if err != nil {
				return nil, errors.Errorf("globbing: %w", err)
			}
			if matches {
				fmtr, ok := getFormatterByLanguage(ctx, fmt)
				if !ok {
					// should never happen
					panic("unknown formatter: " + fmt)
				}
				return fmtr, nil
			}
		}

		return nil, errors.New("no formatter found for file at path: " + filename)
	}

	fmtr, ok := getFormatterByLanguage(ctx, formatter)
	if !ok {
		return nil, errors.New("unknown formatter name: " + formatter)
	}
	return fmtr, nil
}

func getFormatterByLanguage(ctx context.Context, lang string) (format.Provider, bool) {
	switch strings.ToLower(lang) {
	case "hcl", "hcl2", "terraform", "tf":
		return hclfmt.NewFormatter(), true
	case "proto", "protobuf", "proto3", "protocol buffer":
		return protofmt.NewFormatter(), true
	case "yaml", "yml":
		return yamlfmt.NewFormatter(), true
	case "sh", "bash", "zsh", "ksh", "shell":
		return shfmt.NewFormatter(), true
	case "dockerfile", "docker":
		return dockerfmt.NewFormatter(), true
	// external formatters
	case "dart":
		return cmdfmt.NewDartFormatter(ctx, "dart"), true
	case "external-terraform":
		return cmdfmt.NewTerraformFormatter(ctx, "terraform"), true
	case "swift":
		return cmdfmt.NewSwiftFormatter(ctx, "/usr/bin/swift"), true
	default:
		return nil, false
	}
}
