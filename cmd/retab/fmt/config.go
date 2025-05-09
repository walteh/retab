package fmt

import (
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters"
	"github.com/walteh/retab/v2/pkg/formatters/cmdfmt"
	"github.com/walteh/retab/v2/pkg/formatters/dartfmt"
	"github.com/walteh/retab/v2/pkg/formatters/dockerfmt"
	"github.com/walteh/retab/v2/pkg/formatters/hclfmt"
	"github.com/walteh/retab/v2/pkg/formatters/protofmt"
	"github.com/walteh/retab/v2/pkg/formatters/shfmt"
	"github.com/walteh/retab/v2/pkg/formatters/swiftfmt"
	"github.com/walteh/retab/v2/pkg/formatters/terraformfmt"
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
		DartFmt:      format.NewLazyFormatProvider(func() format.Provider { return dartfmt.NewDartCmdFormatter(cmdfmt.WithUseDocker(true)) }),
		TerraformFmt: format.NewLazyFormatProvider(func() format.Provider { return terraformfmt.NewTerraformCmdFormatter(cmdfmt.WithUseDocker(true)) }),
		SwiftFmt:     format.NewLazyFormatProvider(func() format.Provider { return swiftfmt.NewSwiftCmdFormatter(cmdfmt.WithUseDocker(true)) }),
	}

	return cfg
}
