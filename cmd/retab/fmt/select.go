package fmt

import (
	"context"

	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
	"github.com/walteh/retab/v2/pkg/format/hclfmt"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
	"github.com/walteh/retab/v2/pkg/format/yamlfmt"
	"gitlab.com/tozd/go/errors"
)

func getFormatter(ctx context.Context, formatter string, filename string) (format.Provider, error) {
	if formatter == "auto" {
		formatters := []format.Provider{
			hclfmt.NewFormatter(),
			protofmt.NewFormatter(),
			cmdfmt.NewDartFormatter("dart"),
			cmdfmt.NewTerraformFormatter("terraform"),
			cmdfmt.NewSwiftFormatter("swift"),
			yamlfmt.NewFormatter(),
		}
		fmtr, err := format.AutoDetectFormatter(filename, formatters)
		if err != nil {
			return nil, errors.Errorf("auto-detecting formatter: %w", err)
		}
		if fmtr == nil {
			return nil, errors.Errorf("no formatters found for file '%s'", filename)
		}
		return fmtr, nil
	}

	switch formatter {
	case "hcl":
		return hclfmt.NewFormatter(), nil
	case "proto":
		return protofmt.NewFormatter(), nil
	case "dart":
		return cmdfmt.NewDartFormatter("dart"), nil
	case "tf":
		return cmdfmt.NewTerraformFormatter("terraform"), nil
	case "swift":
		return cmdfmt.NewSwiftFormatter("swift"), nil
	case "yaml":
		return yamlfmt.NewFormatter(), nil
	default:
		return nil, errors.New("invalid formatter")
	}
}
