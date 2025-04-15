package fmt

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-enry/go-enry/v2"
	"github.com/rs/zerolog"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/dockerfmt"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"github.com/walteh/retab/v2/pkg/format/shfmt"
	"github.com/walteh/retab/v2/pkg/format/yamlfmt"
	"gitlab.com/tozd/go/errors"
)

func trackStats(ctx context.Context) func() {

	memoryStart := runtime.MemStats{}
	runtime.ReadMemStats(&memoryStart)

	// Track goroutine count
	goroutinesStart := runtime.NumGoroutine()

	// Track GC stats
	gcStart := memoryStart.NumGC

	start := time.Now()

	return func() {
		duration := time.Since(start)

		memoryEnd := runtime.MemStats{}
		runtime.ReadMemStats(&memoryEnd)

		// Get final goroutine count
		goroutinesEnd := runtime.NumGoroutine()

		// Calculate additional metrics
		memoryUsage := memoryEnd.TotalAlloc - memoryStart.TotalAlloc
		gcRuns := memoryEnd.NumGC - gcStart

		zerolog.Ctx(ctx).Info().
			Str("duration", duration.String()).
			Uint64("memory_usage_bytes", memoryUsage).
			Str("memory_usage_human", humanizeBytes(memoryUsage)).
			Int("goroutines", goroutinesEnd-goroutinesStart).
			Uint32("gc_runs", gcRuns).
			Float64("gc_pause_total_ms", float64(memoryEnd.PauseTotalNs-memoryStart.PauseTotalNs)/1000000).
			Msg("fmt completed")
	}
}

// Helper function to make byte sizes human-readable
func humanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
				zerolog.Ctx(ctx).Warn().Str("filename", filename).Str("glob", glob).Msg("unknown formatter - should never happen")
				return nil, false
			}
			zerolog.Ctx(ctx).Info().Str("filename", filename).Str("formatter", fmt).Msg("detected formatter (fast)")
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

		zerolog.Ctx(ctx).Info().Str("filename", filename).Str("language_detected", lang).Msg("detected formatter (fallback)")

		return fmtr, true
	}

	zerolog.Ctx(ctx).Warn().Str("filename", filename).Strs("languages_detected", langs).Msg("fallback:no formatter found for detected languages")

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

		return nil, errors.New("unable to auto-detect formatter for file at path: " + filename)
	}

	fmtr, ok := getFormatterByLanguage(ctx, formatter)
	if !ok {
		return nil, errors.Errorf("unknown formatter name [%s] - trying to format file at path: %s", formatter, filename)
	}
	return fmtr, nil
}
