package root

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/walteh/retab/cmd/root/fmt"
	"github.com/walteh/retab/cmd/root/gen"
	"github.com/walteh/retab/cmd/root/resolvers"
	"github.com/walteh/retab/cmd/root/validate"
	"github.com/walteh/retab/cmd/root/wfmt"
	"github.com/walteh/snake"
	"github.com/walteh/snake/scobra"
)

func NewCommand(ctx context.Context) (*scobra.CobraSnake, *cobra.Command, error) {

	cmd := &cobra.Command{
		Use:   "retab",
		Short: "retab brings tabs to your code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	impl := scobra.NewCobraSnake(ctx, cmd)

	opts := snake.Opts(
		snake.Commands(
			scobra.NewCommand(&wfmt.Handler{}),
			scobra.NewCommand(&fmt.Handler{}),
			scobra.NewCommand(&gen.Handler{}),
			scobra.NewCommand(&validate.Handler{}),
		),
		snake.Resolvers(
			resolvers.FSRunner(),
			resolvers.ConfigurationRunner(),
		),
	)
	_, err := snake.NewSnakeWithOpts(ctx, impl, opts)
	if err != nil {
		return nil, nil, err
	}

	return impl, cmd, nil
}
