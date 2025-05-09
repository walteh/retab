package formatters

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
)

type AutoFormatProvider struct {
	HCLFmt       format.Provider
	ProtoFmt     format.Provider
	YAMLFmt      format.Provider
	ShFmt        format.Provider
	DockerFmt    format.Provider
	DartFmt      format.Provider
	TerraformFmt format.Provider
	SwiftFmt     format.Provider
}

type LanguageConfig struct {
	LangIds       []string
	FilenameGlobs []string
	ProviderFunc  func(me *AutoFormatProvider) format.Provider
}

var (
	hclConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"hcl", "hcl2", "terraform", "tf"},
		FilenameGlobs: []string{"*.{hcl,hcl2,terraform,tf,tfvars}"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.HCLFmt },
	})
	protoConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"proto", "proto3"},
		FilenameGlobs: []string{"*.proto"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.ProtoFmt },
	})
	yamlConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"yaml", "yml"},
		FilenameGlobs: []string{"*.yaml", "*.yml"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.YAMLFmt },
	})
	shConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"sh", "bash", "zsh", "ksh", "shell"},
		FilenameGlobs: []string{"*.sh", "*.bash", "*.zsh", "*.ksh", "*.shell"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.ShFmt },
	})
	dockerConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"dockerfile", "docker"},
		FilenameGlobs: []string{"Dockerfile", "Dockerfile.*"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.DockerFmt },
	})
	dartConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"dart"},
		FilenameGlobs: []string{"*.dart"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.DartFmt },
	})
	terraformConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"external-terraform"},
		FilenameGlobs: []string{},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.TerraformFmt },
	})
	swiftConfig = RegisterLanguageConfig(&LanguageConfig{
		LangIds:       []string{"swift"},
		FilenameGlobs: []string{"*.swift"},
		ProviderFunc:  func(me *AutoFormatProvider) format.Provider { return me.SwiftFmt },
	})
)

func (me *AutoFormatProvider) GetFormatter(ctx context.Context, formatter string, filename string, content io.ReadSeeker) (format.Provider, error) {

	if formatter == "auto" || formatter == "" {

		fmtr, ok := me.DetectFormatterFromFilenameGlobs(ctx, filename)
		if ok {
			return fmtr, nil
		}

		fmtr, ok = me.DetectFormatterFromContent(ctx, filename, content)
		if ok {
			return fmtr, nil
		}

		return nil, oops.WithContext(ctx).Errorf("unable to auto-detect formatter")
	}

	fmtr, ok := me.GetFormatterByLangID(ctx, formatter)
	if !ok {
		return nil, oops.WithContext(ctx).With("formatter_arg", formatter).Errorf("unknown formatter name")
	}
	return fmtr, nil
}

var languageConfigs = []*LanguageConfig{}

func RegisterLanguageConfig(config *LanguageConfig) *LanguageConfig {
	languageConfigs = append(languageConfigs, config)
	return config
}

func (me *AutoFormatProvider) GetFormatterByLangID(ctx context.Context, lang string) (format.Provider, bool) {
	lang = strings.ToLower(lang)
	for _, config := range languageConfigs {
		for _, langId := range config.LangIds {
			if langId == lang {
				return config.ProviderFunc(me), true
			}
		}
	}

	return nil, false
}

func (me *AutoFormatProvider) DetectFormatterFromFilenameGlobs(ctx context.Context, filename string) (format.Provider, bool) {

	for _, config := range languageConfigs {
		for _, glob := range config.FilenameGlobs {
			matches, err := doublestar.PathMatch(glob, filepath.Base(filename))
			if err != nil {
				// should never happen
				panic("globbing: " + err.Error())
			}
			if matches {
				provider := config.ProviderFunc(me)
				zerolog.Ctx(ctx).Info().Str("glob", glob).Type("detected_formatter", provider).Msg("detected formatter (fast)")
				return provider, true
			}
		}
	}

	return nil, false
}

func (me *AutoFormatProvider) DetectFormatterFromContent(ctx context.Context, filename string, br io.ReadSeeker) (format.Provider, bool) {

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
		fmtr, ok := me.GetFormatterByLangID(ctx, lang)
		if !ok {
			continue
		}

		zerolog.Ctx(ctx).Info().Str("language_detected", lang).Msg("detected formatter (fallback)")

		return fmtr, true
	}

	zerolog.Ctx(ctx).Warn().Strs("languages_detected", langs).Msg("fallback:no formatter found for detected languages")

	return nil, false
}
