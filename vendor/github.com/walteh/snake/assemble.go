package snake

import (
	"context"
	"reflect"

	"github.com/spf13/cobra"
)

func Apply(ctx context.Context, r *cobra.Command) error {
	return ApplyCtx(ctx, &root, r)
}

func ApplyCtx(ctx context.Context, me *Ctx, root *cobra.Command) error {

	if root.RunE == nil {
		root.RunE = func(cmd *cobra.Command, args []string) error {
			return nil
		}
	}

	for k, exer := range me.resolvers {

		if exer.Command() == nil {
			continue
		}

		cmd := exer.Command().Cobra()

		if flgs, err := me.FlagsFor(exer); err != nil {
			return err
		} else {
			cmd.Flags().AddFlagSet(flgs)
		}

		err := exer.ValidateResponse()
		if err != nil {
			return err
		}

		oldRunE := cmd.RunE

		// hold a reference to the current value of k
		holdk := k

		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			defer setBindingWithLock(me, cmd)()
			defer setBindingWithLock(me, args)()

			err := runResolvingArguments(holdk, func(s string) IsRunnable {
				return me.resolvers[s]
			}, me.bindings)
			if err != nil {
				return err
			}
			if oldRunE != nil {
				return oldRunE(cmd, args)
			}
			return nil
		}

		root.AddCommand(cmd)

	}

	return nil
}

func Build(ctx context.Context) (*cobra.Command, error) {
	return BuildCtx(ctx, &root)
}

func BuildCtx(ctx context.Context, me *Ctx) (*cobra.Command, error) {

	cmd := &cobra.Command{}

	if err := ApplyCtx(ctx, me, cmd); err != nil {
		return nil, err
	}

	for nme, sub := range me.cmds {
		cmdn := sub.Cobra()
		cmd.AddCommand(cmdn)

		if cmdn.Use == "" {
			cmdn.Use = nme
		}
	}

	return cmd, nil

}

var end_of_chain = reflect.ValueOf("end_of_chain")
var end_of_chain_ptr = &end_of_chain
