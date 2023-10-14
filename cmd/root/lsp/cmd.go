package lsp

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/walteh/snake"
)

var _ snake.Cobrad = (*Handler)(nil)

type Handler struct {
}

func (me *Handler) Cobra() *cobra.Command {
	cmd := &cobra.Command{
		Short: "run a server for retab code using the Language Server Protocol",
	}

	cmd.Args = cobra.ExactArgs(0)

	return cmd
}

func (me *Handler) Run(ctx context.Context) error {
	return nil
	// return NewServe().Run(debug.WithInstance(ctx, "./de.bug", "serve"), nil)
}
