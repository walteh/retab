package fmt

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-enry/go-enry/v2"
	"github.com/rs/zerolog"
	"github.com/samber/oops"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/dockerfmt"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"github.com/walteh/retab/v2/pkg/format/shfmt"
	"github.com/walteh/retab/v2/pkg/format/yamlfmt"
)

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
		return cmdfmt.NewSwiftFormatter(ctx, "swift"), true
	default:
		return nil, false
	}
}

func autoDetectFast(ctx context.Context, filename string) (format.Provider, bool) {
	var backupAutoDetectFormatterGlobs = map[string]string{
		"*.{hcl,hcl2,terraform,tf,tfvars}": "hcl",
		"*.{proto,proto3}":                 "proto",
		"*.dart":                           "dart",
		"*.swift":                          "swift",
		"*.{yaml,yml}":                     "yaml",
		"*.{sh,bash,zsh,ksh,shell}":        "sh",
		// ".{bash,zsh}{rc,_history}":         "sh", // commenting for now just to make sure the fallback works
		"{Dockerfile,Dockerfile.*}": "dockerfile",
	}

	for glob, fmt := range backupAutoDetectFormatterGlobs {
		matches, err := doublestar.PathMatch(glob, filepath.Base(filename))
		if err != nil {
			// should never happen
			panic("globbing: " + err.Error())
		}
		if matches {
			fmtr, ok := getFormatterByLanguage(ctx, fmt)
			if !ok {
				zerolog.Ctx(ctx).Warn().Str("glob", glob).Msg("unknown formatter - should never happen")
				return nil, false
			}
			zerolog.Ctx(ctx).Info().Str("glob", glob).Type("detected_formatter", fmtr).Msg("detected formatter (fast)")
			return fmtr, true
		}
	}

	return nil, false
}

func autoDetectFallback(ctx context.Context, filename string, br io.ReadSeeker) (format.Provider, bool) {

	// Peek at the first 250 bytes without advancing the reader
	peeked := make([]byte, 250)
	_, _ = br.Read(peeked)
	defer br.Seek(0, io.SeekStart)

	langs := enry.GetLanguages(filename, peeked)
	if len(langs) == 0 {
		return nil, false
	}

	zerolog.Ctx(ctx).Debug().Strs("languages", langs).Msg("found languages")

	for _, lang := range langs {
		fmtr, ok := getFormatterByLanguage(ctx, lang)
		if !ok {
			continue
		}

		zerolog.Ctx(ctx).Info().Str("language_detected", lang).Msg("detected formatter (fallback)")

		return fmtr, true
	}

	zerolog.Ctx(ctx).Warn().Strs("languages_detected", langs).Msg("fallback:no formatter found for detected languages")

	return nil, false
}

func getFormatter(ctx context.Context, formatter string, filename string, br io.ReadSeeker) (format.Provider, error) {

	if formatter == "auto" || formatter == "" {

		fmtr, ok := autoDetectFast(ctx, filename)
		if ok {
			return fmtr, nil
		}

		fmtr, ok = autoDetectFallback(ctx, filename, br)
		if ok {
			return fmtr, nil
		}

		return nil, oops.WithContext(ctx).Errorf("unable to auto-detect formatter")
	}

	fmtr, ok := getFormatterByLanguage(ctx, formatter)
	if !ok {
		return nil, oops.WithContext(ctx).With("formatter_arg", formatter).Errorf("unknown formatter name", filename)
	}
	return fmtr, nil
}
