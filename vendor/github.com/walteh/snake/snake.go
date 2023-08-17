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
	ErrMissingBinding = fmt.Errorf("snake.ErrMissingBinding")
	ErrMissingRun     = fmt.Errorf("snake.ErrMissingRun")
	ErrInvalidRun     = fmt.Errorf("snake.ErrInvalidRun")
)

func NewRootCommand(ctx context.Context, snk Snakeable) *cobra.Command {

	cmd := snk.BuildCommand(ctx)

	cmd.PersistentPreRunE = func(ccc *cobra.Command, args []string) error {
		if err := cmd.ParseFlags(args); err != nil {
			return err
		}

		err := snk.ParseArguments(ccc.Context(), ccc, args)
		if err != nil {
			return err
		}

		return nil
	}

	cmd.RunE = func(ccc *cobra.Command, args []string) error {
		err := snk.ParseArguments(ccc.Context(), ccc, args)
		if err != nil {
			return err
		}
		return nil
	}

	cmd.SetContext(ctx)

	return cmd

}

func MustNewCommand(ctx context.Context, cbra *cobra.Command, snk Snakeable) {
	err := NewCommand(ctx, cbra, snk)
	if err != nil {
		panic(err)
	}
}

func NewCommand(ctx context.Context, cbra *cobra.Command, snk Snakeable) error {

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
		return callRunMethod(cmd.Context(), method, tpe)
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
	if tk.Kind() == reflect.Ptr {
		tk = tk.Elem()
	}

	pre[tk] = func() (reflect.Value, error) { return reflect.ValueOf(value), nil }

	return context.WithValue(ctx, &bindingsKeyT{}, pre)
}

type bindings map[reflect.Type]func() (reflect.Value, error)

type bindingsKeyT struct {
}

var callbackReturnSignature = reflect.TypeOf((*error)(nil)).Elem()

func callRunMethod(ctx context.Context, f reflect.Value, t reflect.Type) error {

	in := []reflect.Value{}

	for i := 0; i < t.NumIn(); i++ {
		pt := t.In(i)
		if pt.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			in = append(in, reflect.ValueOf(ctx))
		} else {
			b, ok := ctx.Value(&bindingsKeyT{}).(bindings)
			if !ok {
				return errors.WithMessage(ErrMissingBinding, "no snake bindings in context")
			}
			bv, ok := b[pt]
			if !ok {
				return errors.WithMessage(ErrMissingBinding, fmt.Sprintf("no binding for %s", pt))
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

	// either no arguments or first argument must be context.Context
	if t.NumIn() != 0 && !t.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return nil, errors.WithMessagef(ErrInvalidRun, "first argument of (%s).Run must be of type \"context.Context\"", parentName)
	}

	// must return only an error to comply with cobra.Command.RunE
	if t.NumOut() != 1 || !t.Out(0).Implements(callbackReturnSignature) {
		return nil, errors.WithMessagef(ErrInvalidRun, "return value of (%s).Run must be of type \"error\"", parentName)
	}

	return t, nil
}
