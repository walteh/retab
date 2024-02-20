package scobra

import (
	"context"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/walteh/snake"
	"github.com/walteh/terrors"
)

var (
	_ snake.SnakeImplementationTyped[*cobra.Command] = &CobraSnake{}
)

type CobraSnake struct {
	RootCommand *cobra.Command
}

func NewCommandResolver(s *cobra.Command) snake.TypedResolver[*cobra.Command] {
	return snake.MustGetTypedResolver(s)
}

func (me *CobraSnake) ManagedResolvers(_ context.Context) []snake.UntypedResolver {
	return []snake.UntypedResolver{
		snake.NewNoopMethod[*cobra.Command](),
		snake.NewNoopMethod[[]string](),
	}
}

func applyInputToFlags(input snake.Input, flgs *pflag.FlagSet) error {
	switch t := input.(type) {
	case *snake.StringEnumInput:
		flgs.Var(NewWrappedEnum(t), input.Name(), t.Usage())
	case *snake.StringInput:
		flgs.StringVar(t.Value(), input.Name(), t.Default(), t.Usage())
	case *snake.BoolInput:
		flgs.BoolVar(t.Value(), input.Name(), t.Default(), t.Usage())
	case *snake.IntInput:
		flgs.IntVar(t.Value(), input.Name(), t.Default(), t.Usage())
	case *snake.StringArrayInput:
		flgs.StringSliceVar(t.Value(), input.Name(), t.Default(), t.Usage())
	case *snake.IntArrayInput:
		flgs.IntSliceVar(t.Value(), input.Name(), t.Default(), t.Usage())
	case *snake.DurationInput:
		flgs.DurationVar(t.Value(), input.Name(), t.Default(), t.Usage())
	default:
		return terrors.Errorf("unknown input type %T", t)
	}
	return nil
}

func (me *CobraSnake) Decorate(ctx context.Context, self snake.TypedResolver[*cobra.Command], snk snake.Snake, inputs []snake.Input, mw []snake.Middleware) error {

	cmd := self.TypedRef()

	if cmd.Use == "" {
		cmd.Use = self.Name()
	}

	if cmd.Short == "" {
		cmd.Short = self.Description()
	}

	name := cmd.Name()

	oldRunE := cmd.RunE

	for _, v := range inputs {
		flgs := cmd.Flags()

		if v.Shared() {
			flgs = cmd.PersistentFlags()
		}
		if flgs.Lookup(v.Name()) != nil {
			// if this is the same object, then the user is trying to override the flag, so we let them
			// or (more likely) this is a shared flag, so we are setting it over and over again
			continue
		}

		err := applyInputToFlags(v, flgs)
		if err != nil {
			return err
		}

		if v.Shared() {
			cmd.PersistentFlags().Lookup(v.Name()).Shorthand = ""
			continue
		}
	}

	// if a flag is not set, we check the environment for "cmd_name_arg_name"
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Changed {
				return
			}
			val := strings.ToUpper(me.RootCommand.Name() + "_" + strings.ReplaceAll(f.Name, "-", "_"))
			envvar := os.Getenv(val)
			if envvar == "" {
				return
			}
			auto, err := snake.AutoENVVar(ctx, envvar)
			if err != nil {
				panic(err)
			}
			err = f.Value.Set(auto)
			if err != nil {
				panic(err)
			}
		})
		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		binder := snake.NewBinder()

		snake.SetBinding(binder, cmd)
		snake.SetBinding(binder, args)

		outhand := NewOutputHandler(cmd)

		err := snake.RunResolvingArguments(outhand, snk.Resolve, name, binder, mw...)
		if err != nil {
			return err
		}
		if oldRunE != nil {
			err := oldRunE(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}

	me.RootCommand.AddCommand(cmd)

	return nil
}

func (me *CobraSnake) OnSnakeInit(ctx context.Context, snk snake.Snake) error {

	me.RootCommand.RunE = func(cmd *cobra.Command, args []string) error {
		binder := snake.NewBinder()

		snake.SetBinding(binder, cmd)
		snake.SetBinding(binder, args)

		outhand := NewOutputHandler(cmd)

		err := snake.RunResolvingArguments(outhand, snk.Resolve, "root", binder)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

var _ snake.EnumResolverFunc = (*CobraSnake)(nil).ResolveEnum

func (me *CobraSnake) ResolveEnum(typ string, opts []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select " + typ,
		Items: opts,
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	if result == "" {
		return "", terrors.Errorf("invalid %q", typ)
	}

	return result, nil
}

func (me *CobraSnake) ProvideContextResolver() snake.UntypedResolver {
	return snake.MustGetResolverFor[context.Context](&ContextResolver{})
}

func NewCobraSnake(ctx context.Context, root *cobra.Command) *CobraSnake {

	if root == nil {
		root = &cobra.Command{}
	}

	me := &CobraSnake{root}

	str, err := DecorateTemplate(ctx, root, &DecorateOptions{
		Headings:      color.New(color.FgHiCyan, color.Bold),
		ExecName:      color.New(color.FgHiGreen, color.Bold),
		Commands:      color.New(color.Bold, color.FgGreen),
		FlagsDataType: color.New(color.Faint),
		Flags:         color.New(color.Bold),
	})
	if err != nil {
		panic(err)
	}

	root.SetUsageTemplate(str)

	root.SilenceUsage = true

	return me
}

func ExecuteHandlingError(ctx context.Context, cmd *CobraSnake) {
	cmd.RootCommand.SilenceErrors = true
	err := HandleErrorByPrintingToConsole(cmd.RootCommand, cmd.RootCommand.ExecuteContext(ctx))
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

type Cobrad interface {
	snake.RegisterableRunFunc
	CobraCommand() *cobra.Command
}

func NewCommand(f Cobrad) snake.TypedResolver[*cobra.Command] {
	cmd := f.CobraCommand()
	if cmd.Use == "" {
		cmd.Use = cmd.Name()
	}
	return snake.NewInlineNamedRunner(cmd, f, cmd.Name(), cmd.Short)
}
