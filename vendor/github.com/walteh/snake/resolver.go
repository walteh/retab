package snake

import (
	"context"
	"reflect"

	"github.com/go-faster/errors"
)

func ResolveBindingsFromProvider(ctx context.Context, rf reflect.Value) (context.Context, error) {

	loa := listOfArgs(rf.Type())

	// add a unique context value to valitate the context resolver is returning a child context
	// will only be used by this function
	type contextResolverKeyT struct {
	}
	contextResolverKey := &contextResolverKeyT{}
	ctx = context.WithValue(ctx, contextResolverKey, true)

	if len(loa) > 1 {
		for i, pt := range loa {
			if i == 0 {
				continue
			}
			if pt.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
				// move the context to the first position
				// this is so that subsequent bindings can be resolved with the updated context
				loa[0], loa[i] = loa[i], loa[0]
				break
			}
		}
	}

	for _, pt := range loa {
		rslv, ok := getResolvers(ctx)[pt.String()]
		if !ok {
			if pt.Kind() == reflect.Ptr {
				pt = pt.Elem()
			}
			// check if we have a flag binding for this type
			rslv, ok = getResolvers(ctx)[pt.String()]
			if !ok {
				continue
			}
		}

		args := []reflect.Value{}
		for _, pt := range listOfArgs(reflect.TypeOf(rslv)) {
			if pt == reflect.TypeOf((*context.Context)(nil)).Elem() {
				args = append(args, reflect.ValueOf(ctx))
			} else {
				// if we end up here, we need to validate the bindings exist
				b, ok := ctx.Value(&bindingsKeyT{}).(bindings)
				if !ok {
					return ctx, errors.Wrapf(ErrMissingBinding, "no snake bindings in context, looking for type %q", pt)
				}
				if _, ok := b[pt.String()]; !ok {
					return ctx, errors.Wrapf(ErrMissingBinding, "no snake binding for type %q", pt)
				}
				args = append(args, reflect.ValueOf(b[pt.String()]))
			}
		}

		p, err := rslv(ctx)
		if err != nil {
			return ctx, err
		}

		if p.IsZero() {
			return ctx, errors.Wrapf(ErrInvalidResolver, "resolver for type %q returned a zero value", pt)
		}

		if reflect.TypeOf(p.Interface()).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {

			// if the provider returns a context - meaning the dyanmic context binding resolver was set
			// we need to merge any bindings that might have been set
			crb := p.Interface().(context.Context)

			// check if the context resolver returned a child context
			if _, ok := crb.Value(contextResolverKey).(bool); !ok {
				return ctx, ErrInvalidResolver
			}

			// this is the context resolver binding, we need to process it
			// we favor the context returned from the resolver, as it also might have been modified
			// for example, if the resolver returns a context with a zerolog logger, we want to keep that
			ctx = mergeBindingKeepingFirst(crb, ctx)

		}
		ctx = bindRaw(ctx, pt, p)

	}

	return ctx, nil
}

func setResolvers(ctx context.Context, provider resolvers) context.Context {
	return context.WithValue(ctx, &resolverKeyT{}, provider)
}

func getResolvers(ctx context.Context) resolvers {
	p, ok := ctx.Value(&resolverKeyT{}).(resolvers)
	if ok {
		return p
	}
	return resolvers{}
}

func setFlagBindings(ctx context.Context, provider flagbindings) context.Context {
	return context.WithValue(ctx, &flagbindingsKeyT{}, provider)
}

func getFlagBindings(ctx context.Context) flagbindings {
	p, ok := ctx.Value(&flagbindingsKeyT{}).(flagbindings)
	if ok {
		return p
	}
	return flagbindings{}
}

func RegisterBindingResolver[I any](ctx context.Context, res typedResolver[I], f ...flagbinding) context.Context {
	// check if we have a dynamic binding resolver available
	dy := getResolvers(ctx)

	elm := reflect.TypeOf((*I)(nil)).Elem()

	dy[elm.String()] = res.asResolver()

	ctx = setResolvers(ctx, dy)

	for _, fbb := range f {
		fb := getFlagBindings(ctx)
		fb[elm.String()] = fbb
		ctx = setFlagBindings(ctx, fb)
	}

	return ctx
}
