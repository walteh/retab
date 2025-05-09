package terraformfmt

import (
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters/cmdfmt"
)

func rawTerraformCmd() []string {
	return []string{"fmt", "-write=false", "-list=false"}
}

func NewTerraformCmdFormatter(opts ...cmdfmt.OptBasicExternalFormatterOptsSetter) format.Provider {
	cmds := rawTerraformCmd()

	startopts := []cmdfmt.OptBasicExternalFormatterOptsSetter{
		cmdfmt.WithIndent("  "),
		cmdfmt.WithExecutable("terraform"),
		cmdfmt.WithDockerImageName("hashicorp/terraform"),
		cmdfmt.WithDockerImageTag("latest"),
	}

	return cmdfmt.NewFormatter(cmds, append(startopts, opts...)...)
}
