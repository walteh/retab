package snake

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func bindRaw(ctx context.Context, key reflect.Type, value reflect.Value) context.Context {
	pre, ok := ctx.Value(&bindingsKeyT{}).(bindings)
	if !ok {
		pre = bindings{}
	}

	pre[key.String()] = func() (reflect.Value, error) { return value, nil }

	return context.WithValue(ctx, &bindingsKeyT{}, pre)
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

	pre[tk.String()] = func() (reflect.Value, error) { return reflect.ValueOf(value), nil }

	return context.WithValue(ctx, &bindingsKeyT{}, pre)
}

func BindG[T any](ctx context.Context, value T) context.Context {
	return bindRaw(ctx, reflect.TypeOf((*T)(nil)).Elem(), reflect.ValueOf(value))
}

type binding func() (reflect.Value, error)

type bindings map[string]binding

type bindingsKeyT struct {
}

func (me typedResolver[T]) asResolver() resolver {
	return func(ctx context.Context) (reflect.Value, error) {
		v, err := me(ctx)
		return reflect.ValueOf(v), err
	}
}

type untypedResolver any

func untypedResolverToResolver(me untypedResolver) resolver {
	return func(ctx context.Context) (reflect.Value, error) {
		v, err := me.(typedResolver[any])(ctx)
		return reflect.ValueOf(v), err
	}
}

type typedResolver[T any] func(context.Context) (T, error)

type resolver func(context.Context) (reflect.Value, error)
type resolvers map[string]resolver
type resolverKeyT struct {
}

type flagbinding func(*pflag.FlagSet)
type flagbindings map[string]flagbinding
type flagbindingsKeyT struct {
}

var callbackReturnSignature = reflect.TypeOf((*error)(nil)).Elem()

func mergeBindingKeepingFirst(to context.Context, from context.Context) context.Context {
	t, ok := to.Value(&bindingsKeyT{}).(bindings)
	if !ok {
		t = bindings{}
	}

	f, ok := from.Value(&bindingsKeyT{}).(bindings)
	if !ok {
		return to
	}

	for k, v := range f {
		_, ok := t[k]
		if !ok {
			t[k] = v
		}
	}

	return context.WithValue(to, &bindingsKeyT{}, t)
}

func mergeResolversKeepingFirst(to context.Context, from context.Context) context.Context {
	f, ok := from.Value(&resolverKeyT{}).(resolvers)
	if !ok {
		return to
	}

	t, ok := to.Value(&resolverKeyT{}).(resolvers)
	if !ok {
		return to
	}

	for k, v := range f {
		_, ok := t[k]
		if !ok {
			t[k] = v
		}
	}

	return context.WithValue(to, &resolverKeyT{}, t)
}

func listOfArgs(typ reflect.Type) []reflect.Type {
	var args []reflect.Type

	for i := 0; i < typ.NumIn(); i++ {
		args = append(args, typ.In(i))
	}

	return args
}

func listOfReturns(typ reflect.Type) []reflect.Type {
	var args []reflect.Type

	for i := 0; i < typ.NumOut(); i++ {
		args = append(args, typ.Out(i))
	}

	return args
}

func callRunMethod(cmd *cobra.Command, f reflect.Value, t reflect.Type) error {

	in := []reflect.Value{}

	// we do not check for the existence of the bindings key here
	// because it might not be needed
	b, bindingsExist := cmd.Context().Value(&bindingsKeyT{}).(bindings)

	contextOverrideExists := false
	if bindingsExist {
		_, ok := b[reflect.TypeOf((*context.Context)(nil)).Elem().String()]
		if ok {
			contextOverrideExists = true
		}
	}

	for _, pt := range listOfArgs(t) {

		if !contextOverrideExists && pt.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			in = append(in, reflect.ValueOf(cmd.Context()))
		} else if pt == reflect.TypeOf((*cobra.Command)(nil)) {
			in = append(in, reflect.ValueOf(cmd))
		} else if pt == reflect.TypeOf(cobra.Command{}) {
			in = append(in, reflect.ValueOf(*cmd))
		} else {
			// if we end up here, we need to validate the bindings exist
			if !bindingsExist {
				return errors.WithMessagef(ErrMissingBinding, "no snake bindings in context, looking for type %q", pt)
			}

			bv, ok := b[pt.String()]
			if !ok {
				return errors.WithMessagef(ErrMissingBinding, "no snake binding for type %q", pt)
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

func commonValidateRunMethod(inter any, method reflect.Value) (reflect.Type, string, error) {

	parentName := reflect.TypeOf(inter).String()

	if method.Kind() == reflect.Invalid || method.IsZero() || method.IsNil() {
		return nil, parentName, errors.WithMessagef(ErrMissingRun, "target ===> %s", parentName)
	}

	if !method.IsValid() {
		return nil, parentName, errors.WithMessagef(ErrInvalidRun, "target ===> %s", parentName)
	}

	if method.Kind() != reflect.Func {
		return nil, parentName, errors.WithMessagef(ErrInvalidRun, "expected function, got %s for (%s).Run", method.Type(), parentName)
	}

	// only here we know it is safe to call Type()
	t := method.Type()

	return t, parentName, nil
}

func validateRunMethod(inter any, method reflect.Value) (reflect.Type, error) {

	t, pname, err := commonValidateRunMethod(inter, method)
	if err != nil {
		return nil, err
	}

	// must return only an error to comply with cobra.Command.RunE
	if t.NumOut() != 1 || !t.Out(0).Implements(callbackReturnSignature) {
		return nil, errors.WithMessagef(ErrInvalidRun, "return value of (%s).Run must be of type \"error\"", pname)
	}

	return t, nil
}

func untypedResovlerReturnSignature[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
