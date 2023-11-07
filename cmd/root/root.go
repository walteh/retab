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
	"github.com/walteh/retab/pkg/configuration"
	"github.com/walteh/snake"
)

func NewCommand() (*cobra.Command, error) {

	cmd := &cobra.Command{
		Use:   "retab",
		Short: "retab brings tabs to terraform",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return snake.NewSnake(&snake.NewSnakeOpts{
		Root: cmd,
		Resolvers: []snake.Method{
			snake.NewArgumentMethod[context.Context](&resolvers.ContextResolver{}),
			snake.NewArgumentMethod[afero.Fs](&resolvers.AferoResolver{}),
			snake.NewArgumentMethod[configuration.Provider](&resolvers.ConfigurationResolver{}),
		},
		Commands: []snake.Method{
			snake.NewCommandMethod(&fmt.Handler{}),
			snake.NewCommandMethod(&buf.Handler{}),
			snake.NewCommandMethod(&hcl.Handler{}),
			snake.NewCommandMethod(&install.Handler{}),
			snake.NewCommandMethod(&generate.Handler{}),
			snake.NewCommandMethod(&lsp.Handler{}),
		},
		GlobalContextResolverFlags: true,
	})

}
