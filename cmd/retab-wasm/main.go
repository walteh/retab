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

	// Create a channel to keep the program running
	c := make(chan struct{}, 0)

	// Register our function
	js.Global().Set("retab_run", js.FuncOf(retabRunWrapper))

	fmt.Println("WASM: Functions registered")
	<-c // Keep the program running
}

func retabRunWrapper(this js.Value, args []js.Value) interface{} {
	if len(args) != 4 {
		js.Global().Set("retab_run_result_error", "Expected 4 arguments: formatter, filename, editorConfigContent, fileContent")
		return false
	}

	formatter := args[0].String()
	filename := args[1].String()
	editorConfigContent := args[2].String()
	fileContent := args[3].String()

	fmt.Printf("WASM: retab_run called with formatter=%s, filename=%s\n", formatter, filename)

	ctx := context.Background()
	result, err := run(ctx, formatter, filename, editorConfigContent, fileContent)
	if err != nil {
		errMsg := err.Error()
		fmt.Printf("WASM: Error in run: %s\n", errMsg)
		js.Global().Set("retab_run_result_error", errMsg)
		return false
	}

	js.Global().Set("retab_run_result_success", result)
	return true
}

func run(ctx context.Context, formatter string, filename string, editorConfigContent string, fileContent string) (result string, err error) {
	fmt.Printf("WASM: Starting run function with formatter=%q\n", formatter)

	var fmtr format.Provider

	cfg, err := editorconfig.NewEditorConfigConfigurationProviderFromContent(ctx, editorConfigContent)
	if err != nil {
		return "", errors.Errorf("failed to create editorconfig configuration provider: %w", err)
	}

	if formatter == "auto" {
		fmt.Println("WASM: Using auto formatter selection")
		pfmters := []format.Provider{}
		pfmters = append(pfmters, hclfmt.NewFormatter())
		pfmters = append(pfmters, protofmt.NewFormatter())
		pfmters = append(pfmters, cmdfmt.NewDartFormatter("dart"))
		pfmters = append(pfmters, cmdfmt.NewTerraformFormatter("terraform"))
		for _, pfmtr := range pfmters {
			for _, target := range pfmtr.Targets() {
				basename := filepath.Base(filename)
				fmt.Printf("WASM: Checking if %s matches pattern %s\n", basename, target)
				ok, err := doublestar.Match(target, basename)
				if err != nil {
					return "", errors.Errorf("failed to match glob: %w", err)
				}
				if ok {
					fmtr = pfmtr
					fmt.Printf("WASM: Found matching formatter for pattern %s\n", target)
					break
				}
			}
		}
	} else if formatter == "hcl" {
		fmt.Println("WASM: Using HCL formatter")
		fmtr = hclfmt.NewFormatter()
	} else if formatter == "proto" {
		fmt.Println("WASM: Using Proto formatter")
		fmtr = protofmt.NewFormatter()
	} else if formatter == "dart" {
		fmt.Println("WASM: Using Dart formatter")
		fmtr = cmdfmt.NewDartFormatter("dart")
	} else if formatter == "tf" {
		fmt.Println("WASM: Using Terraform formatter")
		fmtr = cmdfmt.NewTerraformFormatter("terraform")
	} else {
		fmt.Printf("WASM: Invalid formatter type: %q\n", formatter)
		return "", errors.Errorf("invalid formatter type: %q", formatter)
	}

	if fmtr == nil {
		fmt.Printf("WASM: No formatters found for file %q\n", filename)
		return "", errors.Errorf("no formatters found for file %q", filename)
	}

	fmt.Println("WASM: Formatting content")
	r, err := format.Format(ctx, fmtr, cfg, filename, strings.NewReader(fileContent))
	if err != nil {
		return "", errors.Errorf("failed to format file: %w", err)
	}

	resp, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Errorf("failed to read formatted content: %w", err)
	}

	fmt.Printf("WASM: Formatting complete, result length: %d\n", len(resp))
	return string(resp), nil
}
