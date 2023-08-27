package snake

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Snakeable interface {
	ParseArguments(ctx context.Context, cmd *cobra.Command, args []string) error
	BuildCommand(ctx context.Context) *cobra.Command
}

var (
	ErrMissingBinding   = fmt.Errorf("snake.ErrMissingBinding")
	ErrMissingRun       = fmt.Errorf("snake.ErrMissingRun")
	ErrInvalidRun       = fmt.Errorf("snake.ErrInvalidRun")
	ErrInvalidArguments = fmt.Errorf("snake.ErrInvalidArguments")
)

func NewRootCommand(ctx context.Context, snk Snakeable) *cobra.Command {

	cmd := snk.BuildCommand(ctx)

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := cmd.ParseFlags(args); err != nil {
			return err
		}

		err := snk.ParseArguments(cmd.Context(), cmd, args)
		if err != nil {
			return err
		}

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := snk.ParseArguments(cmd.Context(), cmd, args)
		if err != nil {
			return err
		}
		return nil
	}

	cmd.SetContext(ctx)

	return cmd

}

func NewGroup(ctx context.Context, cmd *cobra.Command, name string, description string) *cobra.Command {

	grp := &cobra.Command{
		Use:   name,
		Short: description,
		Long:  description,
	}

	cmd.AddCommand(grp)

	return cmd
}

func MustNewCommand(ctx context.Context, cbra *cobra.Command, name string, snk Snakeable) {
	err := NewCommand(ctx, cbra, name, snk)
	if err != nil {
		panic(err)
	}
}

func NewCommand(ctx context.Context, cbra *cobra.Command, name string, snk Snakeable) error {

	cmd := snk.BuildCommand(ctx)

	method := getRunMethod(snk)

	tpe, err := validateRunMethod(snk, method)
	if err != nil {
		return err
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		if err := cmd.ParseFlags(args); err != nil {
			return err
		}
		err := snk.ParseArguments(cmd.Context(), cmd, args)
		if err != nil {
			return err
		}
		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return callRunMethod(cmd, method, tpe)
	}

	if name != "" {
		cmd.Use = name
	}

	cbra.AddCommand(cmd)

	return nil
}

func Bind(ctx context.Context, key any, value any) context.Context {
	pre, ok := ctx.Value(&bindingsKeyT{}).(bindings)
	if !ok {
		pre = bindings{}
	}

	tk := reflect.TypeOf(key)
	if tk.Kind() == reflect.Ptr && tk.Elem().Kind() != reflect.Struct {
		tk = tk.Elem()
	}

	pre[tk] = func() (reflect.Value, error) { return reflect.ValueOf(value), nil }

	return context.WithValue(ctx, &bindingsKeyT{}, pre)
}

type bindings map[reflect.Type]func() (reflect.Value, error)

type bindingsKeyT struct {
}

var callbackReturnSignature = reflect.TypeOf((*error)(nil)).Elem()

func callRunMethod(cmd *cobra.Command, f reflect.Value, t reflect.Type) error {

	in := []reflect.Value{}

	// we do not check for the existence of the bindings key here
	// because it might not be needed
	b, bindingsExist := cmd.Context().Value(&bindingsKeyT{}).(bindings)

	contextOverrideExists := false
	if bindingsExist {
		_, ok := b[reflect.TypeOf((*context.Context)(nil)).Elem()]
		if ok {
			contextOverrideExists = true
		}
	}

	for i := 0; i < t.NumIn(); i++ {
		pt := t.In(i)
		if !contextOverrideExists && pt.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			in = append(in, reflect.ValueOf(cmd.Context()))
		} else if pt == reflect.TypeOf((*cobra.Command)(nil)) {
			in = append(in, reflect.ValueOf(cmd))
		} else if pt == reflect.TypeOf(cobra.Command{}) {
			in = append(in, reflect.ValueOf(*cmd))
		} else {
			// if we end up here, we need to validate the bindings exist
			if !bindingsExist {
				return errors.WithMessage(ErrMissingBinding, "no snake bindings in context")
			}

			bv, ok := b[pt]
			if !ok {
				return errors.WithMessagef(ErrMissingBinding, "no snake binding for type %s", pt)
			}

			v, err := bv()
			if err != nil {
				return err
			}

			in = append(in, v)
		}
	}

	out := f.Call(in)
	if out[0].IsNil() {
		return nil
	}
	return out[0].Interface().(error) // nolint
}

func getRunMethod(inter any) reflect.Value {
	value := reflect.ValueOf(inter)
	method := value.MethodByName("Run")
	if !method.IsValid() {
		if value.CanAddr() {
			method = value.Addr().MethodByName("Run")
		}
	}

	return method
}

func validateRunMethod(inter any, method reflect.Value) (reflect.Type, error) {

	parentName := reflect.TypeOf(inter).String()

	if method.Kind() == reflect.Invalid || method.IsZero() || method.IsNil() {
		return nil, errors.WithMessagef(ErrMissingRun, "target ===> %s", parentName)
	}

	if !method.IsValid() {
		return nil, errors.WithMessagef(ErrInvalidRun, "target ===> %s", parentName)
	}

	if method.Kind() != reflect.Func {
		return nil, errors.WithMessagef(ErrInvalidRun, "expected function, got %s for (%s).Run", method.Type(), parentName)
	}

	// only here we know it is safe to call Type()
	t := method.Type()

	// must return only an error to comply with cobra.Command.RunE
	if t.NumOut() != 1 || !t.Out(0).Implements(callbackReturnSignature) {
		return nil, errors.WithMessagef(ErrInvalidRun, "return value of (%s).Run must be of type \"error\"", parentName)
	}

	return t, nil
}
