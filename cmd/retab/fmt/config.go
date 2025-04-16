package fmt

import (
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters"
	"github.com/walteh/retab/v2/pkg/formatters/cmdfmt"
	"github.com/walteh/retab/v2/pkg/formatters/dockerfmt"
	"github.com/walteh/retab/v2/pkg/formatters/hclfmt"
	"github.com/walteh/retab/v2/pkg/formatters/protofmt"
	"github.com/walteh/retab/v2/pkg/formatters/shfmt"
	"github.com/walteh/retab/v2/pkg/formatters/yamlfmt"
)

// currently all formatters are supported by all architectures, if that ever changes we can use this to
// conditionally create the correct formatters
func NewAutoFormatConfig() *formatters.AutoFormatProvider {

	var cfg = &formatters.AutoFormatProvider{
		HCLFmt:       format.NewLazyFormatProvider(func() format.Provider { return hclfmt.NewFormatter() }),
		ProtoFmt:     format.NewLazyFormatProvider(func() format.Provider { return protofmt.NewFormatter() }),
		YAMLFmt:      format.NewLazyFormatProvider(func() format.Provider { return yamlfmt.NewFormatter() }),
		ShFmt:        format.NewLazyFormatProvider(func() format.Provider { return shfmt.NewFormatter() }),
		DockerFmt:    format.NewLazyFormatProvider(func() format.Provider { return dockerfmt.NewFormatter() }),
		DartFmt:      format.NewLazyFormatProvider(func() format.Provider { return cmdfmt.NewDartFormatter("dart") }),
		TerraformFmt: format.NewLazyFormatProvider(func() format.Provider { return cmdfmt.NewTerraformFormatter("terraform") }),
		SwiftFmt:     format.NewLazyFormatProvider(func() format.Provider { return cmdfmt.NewSwiftFormatter("swift") }),
	}

	return cfg
}
