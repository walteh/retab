package snake

import (
	"context"
	"fmt"

	"github.com/walteh/terrors"
)

func RefreshDependencies(rer Input, snk Snake, binder *Binder) (*Binder, error) {

	par := MethodName(rer.Parent())

	deps := snk.DependantsOf(par)

	deps = append(deps, par)

	for _, v := range deps {
		fmt.Printf("deleting binding of %q - is nil: %v \n", v, binder.bindings[v] == nil)
		delete(binder.bindings, v)
	}

	return binder, nil
}

func ResolveAllShared(ctx context.Context, names []string, fmap FMap, binder *Binder) (*Binder, error) {

	for _, v := range names {
		var err error
		var resolver UntypedResolver
		if resolver = fmap(v); resolver == nil {
			return nil, terrors.Errorf("missing resolver for %q", v)
		}

		if MenthodIsShared(resolver) {
			binder, err = findArgumentsRaw(v, fmap, binder)
			if err != nil {
				return nil, err
			}
		}

	}
	return binder, nil
}

func WrapWithMiddleware(base MiddlewareFunc, middlewares ...Middleware) MiddlewareFunc {

	for _, v := range middlewares {
		base = v.Wrap(base)
	}
	return base
}

type Chan chan any

func RunResolvingArguments(outputHandler OutputHandler, fmap FMap, str string, binder *Binder, middlewares ...Middleware) error {
	// always resolve context.Context first
	_, err := findArgumentsRaw("context.Context", fmap, binder)
	if err != nil {
		return err
	}

	base := func(ctx context.Context) error {

		cd := make(Chan)

		defer close(cd)

		closers := make([]func(), 0)

		stdout := outputHandler.Stdout()
		stdin := outputHandler.Stdin()
		stderr := outputHandler.Stderr()

		closers = append(closers, SetBinding[Stdout](binder, stdout))
		closers = append(closers, SetBinding[Stdin](binder, stdin))
		closers = append(closers, SetBinding[Stderr](binder, stderr))
		closers = append(closers, SetBinding[Chan](binder, cd))

		defer func() {
			delete(binder.bindings, str)
			for _, v := range closers {
				v()
			}
		}()

		binder, err := findArgumentsRaw(str, fmap, binder)
		if err != nil {
			return err
		}

		out := binder.bindings[str]

		if out == nil {
			return terrors.Errorf("missing resolver for %q", str)
		}

		result := binder.bindings[str].Interface()

		if out, ok := result.(Output); ok {
			return HandleOutput(ctx, outputHandler, out, cd)
		}

		return nil
	}

	wrp := WrapWithMiddleware(base, middlewares...)

	ctx := binder.bindings["context.Context"].Interface().(context.Context)

	err = wrp(ctx)
	if err != nil {
		return err
	}

	return nil
}
