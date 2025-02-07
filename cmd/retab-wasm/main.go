//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"syscall/js"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/editorconfig"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"gitlab.com/tozd/go/errors"
)

func main() {
	fmt.Println("WASM: Starting main function")
}

//go:wasmexport retab_run
func retab_run(formatter string, filename string, editorConfigContent string, fileContent string) bool {
	ctx := context.Background()
	result, err := run(ctx, formatter, filename, editorConfigContent, fileContent)
	if err != nil {
		js.Global().Set("retab_run_result_error", err.Error())
		return false
	}

	js.Global().Set("retab_run_result_success", result)
	return true
}

func run(ctx context.Context, formatter string, filename string, editorConfigContent string, fileContent string) (result string, err error) {
	fmt.Println("WASM: Starting run function")

	var fmtr format.Provider

	cfg, err := editorconfig.NewEditorConfigConfigurationProviderFromContent(ctx, editorConfigContent)
	if err != nil {
		return "", errors.Errorf("failed to create editorconfig configuration provider: %w", err)
	}

	if formatter == "auto" {
		pfmters := []format.Provider{}
		pfmters = append(pfmters, hclfmt.NewFormatter())
		pfmters = append(pfmters, protofmt.NewFormatter())
		pfmters = append(pfmters, cmdfmt.NewDartFormatter("dart"))
		pfmters = append(pfmters, cmdfmt.NewTerraformFormatter("terraform"))
		for _, pfmtr := range pfmters {
			for _, target := range pfmtr.Targets() {
				basename := filepath.Base(filename)
				ok, err := doublestar.Match(target, basename)
				if err != nil {
					return "", errors.Errorf("failed to match glob: %w", err)
				}
				if ok {
					fmtr = pfmtr
					break
				}
			}
		}
	} else if formatter == "hcl" {
		fmtr = hclfmt.NewFormatter()
	} else if formatter == "proto" {
		fmtr = protofmt.NewFormatter()
	} else if formatter == "dart" {
		fmtr = cmdfmt.NewDartFormatter("dart")
	} else if formatter == "tf" {
		fmtr = cmdfmt.NewTerraformFormatter("terraform")
	} else {
		return "", errors.New("invalid formatter")
	}

	if fmtr == nil {
		return "", errors.Errorf("no formatters found for file '%s'", filename)
	}

	r, err := format.Format(ctx, fmtr, cfg, filename, strings.NewReader(fileContent))
	if err != nil {
		return "", errors.Errorf("failed to format file: %w", err)
	}

	resp, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Errorf("failed to read formatted content: %w", err)
	}

	return string(resp), nil
}
