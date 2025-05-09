package dartfmt

import (
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/formatters/cmdfmt"
)

func rawDartCmd() []string {
	return []string{"format", "--output", "show", "--summary", "none", "--fix"}
}

func NewDartCmdFormatter(opts ...cmdfmt.OptBasicExternalFormatterOptsSetter) format.Provider {
	cmds := rawDartCmd()

	startopts := []cmdfmt.OptBasicExternalFormatterOptsSetter{
		cmdfmt.WithIndent("  "),
		cmdfmt.WithExecutable("dart"),
		cmdfmt.WithDockerImageName("dart"),
		cmdfmt.WithDockerImageTag("stable"),
	}

	return cmdfmt.NewFormatter(cmds, append(startopts, opts...)...)
}
