package cli

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/rs/zerolog"
)

type bindings map[reflect.Type]func() (reflect.Value, error)

type bindingsKeyT struct {
}

var callbackReturnSignature = reflect.TypeOf((*error)(nil)).Elem()

func Bind(ctx context.Context, key any, value any) context.Context {
	pre, ok := ctx.Value(&bindingsKeyT{}).(bindings)
	if !ok {
		pre = bindings{}
	}

	pre[reflect.TypeOf(key).Elem()] = func() (reflect.Value, error) { return reflect.ValueOf(value), nil }

	return context.WithValue(ctx, &bindingsKeyT{}, pre)
}

func callFunction(ctx context.Context, f reflect.Value) error {

	if f.Kind() != reflect.Func {
		return fmt.Errorf("expected function, got %s", f.Type())
	}
	in := []reflect.Value{}
	t := f.Type()
	if t.NumOut() != 1 || !t.Out(0).Implements(callbackReturnSignature) {
		return fmt.Errorf("return value of %s must implement \"error\"", t)
	}
	for i := 0; i < t.NumIn(); i++ {
		pt := t.In(i)
		if pt.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
			in = append(in, reflect.ValueOf(ctx))
		} else {
			b, ok := ctx.Value(&bindingsKeyT{}).(bindings)
			if !ok {
				zerolog.Ctx(ctx).Debug().Msgf("no bindings for %s", pt)
				return errors.New("no bindings")
			}
			bv, ok := b[pt]
			if !ok {
				return fmt.Errorf("no binding for %s", pt)
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

func callMethod(ctx context.Context, name string, v, f reflect.Value) error {
	err := callFunction(ctx, f)
	if err != nil {
		return fmt.Errorf("%s.%s(): %w", v.Type(), name, err)
	}
	return nil
}

func getMethod(inter any, name string) reflect.Value {
	value := reflect.ValueOf(inter)
	method := value.MethodByName(name)
	if !method.IsValid() {
		if value.CanAddr() {
			method = value.Addr().MethodByName(name)
		}
	}
	return method
}
