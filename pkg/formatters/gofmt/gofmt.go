package gofmt

import (
	"bytes"
	"context"
	"io"

	goformat "go/format"

	"github.com/walteh/goimports-reviser/v3/reviser"
	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"

	"strings"
)

type Formatter struct {
}

var _ format.Provider = (*Formatter)(nil)

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (me *Formatter) Format(ctx context.Context, cfg format.Configuration, read io.Reader) (io.Reader, error) {

	raw := cfg.Raw()

	projectName := raw["go_module_name"]
	importRenames := raw["go_rename_imports"]
	renameImportsSeparator := raw["go_rename_imports_separator"]
	yesIWantSpaces := raw["go_yes_i_want_spaces"]
	justFormat := raw["go_just_format"]

	reads, err := io.ReadAll(read)
	if err != nil {
		return nil, errors.Errorf("read: %w", err)
	}

	// fomrat with gofmt
	formattedOutput, err := goformat.Source(reads)
	if err != nil {
		return nil, errors.Errorf("go format: %w", err)
	}

	if justFormat == "true" {
		return bytes.NewReader(formattedOutput), nil
	}

	opts := []reviser.SourceFileOption{
		reviser.WithReader(bytes.NewReader(formattedOutput)),
		reviser.WithImportsOrder([]reviser.ImportsOrder{
			reviser.DottedImportsOrder,
			reviser.BlankedImportsOrder,
			reviser.StdImportsOrder,
			reviser.NamedStdImportsOrder,
			reviser.XImportsOrder,
			reviser.NamedXImportsOrder,
			reviser.GeneralImportsOrder,
			reviser.NamedGeneralImportsOrder,
			reviser.CompanyImportsOrder,
			reviser.NamedCompanyImportsOrder,
			reviser.ProjectImportsOrder,
			reviser.NamedProjectImportsOrder,
		}),
		reviser.WithSeparatedNamedImports,
	}

	if importRenames != "" {
		separator := "="
		if renameImportsSeparator != "" {
			separator = renameImportsSeparator
		}
		for _, rename := range strings.Split(importRenames, ",") {
			parts := strings.Split(rename, separator)
			if len(parts) != 2 {
				return nil, errors.Errorf("invalid rename: %s", rename)
			}
			opts = append(opts, reviser.WithRenameImport(parts[0], parts[1]))
		}
	}

	formattedOutput, originalContent, changed, err := reviser.NewSourceFile(projectName, "").Fix(opts...)
	if err != nil {
		return nil, errors.Errorf("go revise imports: %w", err)
	}

	if !cfg.UseTabs() && yesIWantSpaces == "true" {
		// I really didn't want to do this, because really this formatter is for the imports,
		// and not the code. But I'll give the world the benefit of the doubt and assume that someone
		// has a good enough reason for using spaces in go as I have for using tabs in protobuf.
		ret, err := format.BruteForceIndentation(ctx, "\t", cfg, bytes.NewReader(formattedOutput))
		if err != nil {
			return nil, errors.Errorf("brute force indentation: %w", err)
		}
		return ret, nil
	}

	if !changed {
		return bytes.NewReader(originalContent), nil
	}

	return bytes.NewReader(formattedOutput), nil

}
