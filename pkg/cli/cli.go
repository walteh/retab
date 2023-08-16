package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

type Cobraface interface {
	Inject(ctx context.Context, cmd *cobra.Command, args []string) error
	Define(ctx context.Context) *cobra.Command
}

func RegisterRoot(ctx context.Context, cbrafc Cobraface) *cobra.Command {

	cmd := cbrafc.Define(ctx)

	cmd.PersistentPreRunE = func(ccc *cobra.Command, args []string) error {
		if err := cmd.ParseFlags(args); err != nil {
			return err
		}
		err := cbrafc.Inject(ccc.Context(), ccc, args)
		if err != nil {
			return err
		}

		return nil
	}

	cmd.RunE = func(ccc *cobra.Command, args []string) error {
		err := cbrafc.Inject(ccc.Context(), ccc, args)
		if err != nil {
			return err
		}
		return nil
	}
	cmd.SetContext(ctx)

	return cmd

}

func RegisterCommand(ctx context.Context, cbra *cobra.Command, cbrafc Cobraface) {

	cmd := cbrafc.Define(ctx)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		if err := cmd.ParseFlags(args); err != nil {
			return err
		}

		err := cbrafc.Inject(cmd.Context(), cmd, args)
		if err != nil {
			return err
		}

		method := getMethod(cbrafc, "Run")
		if method.IsValid() {
			return callMethod(cmd.Context(), "Run", method, method)
		}

		return fmt.Errorf("no Run() method found on %T", cbrafc)
	}

	cbra.AddCommand(cmd)

}
