package root

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/buf"
	"github.com/walteh/retab/cmd/root/fmt"
	"github.com/walteh/retab/cmd/root/generate"
	"github.com/walteh/retab/cmd/root/hcl"
	"github.com/walteh/retab/cmd/root/install"
	"github.com/walteh/retab/cmd/root/lsp"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/snake"
)

func NewCommand(ctx context.Context) (*cobra.Command, error) {

	cmd := &cobra.Command{
		Use:   "retab",
		Short: "retab brings tabs to terraform",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	snake.NewCmd(&fmt.Handler{})
	snake.NewCmd(&buf.Handler{})
	snake.NewCmd(&hcl.Handler{})
	snake.NewCmd(&install.Handler{})
	snake.NewCmd(&generate.Handler{})
	snake.NewCmd(&lsp.Handler{})

	snake.NewArgument[context.Context](&resolvers.ContextResolver{})
	snake.NewArgument[afero.Fs](&resolvers.AferoResolver{})

	err := snake.Apply(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
