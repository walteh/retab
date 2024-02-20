package snake

import (
	"context"
	"io"
	"reflect"

	"github.com/walteh/terrors"
)

type NewSnakeOpts[M any] struct {
	Commands             []TypedResolver[M]
	Resolvers            []UntypedResolver
	OverrideEnumResolver EnumResolverFunc
}

func Opts[M any](commands []TypedResolver[M], resolvers []UntypedResolver) *NewSnakeOpts[M] {
	return &NewSnakeOpts[M]{
		Commands:  commands,
		Resolvers: resolvers,
	}
}

type Snake interface {
	ResolverNames() []string
	Resolve(string) UntypedResolver
	Enums() []Enum
	Resolvers() []UntypedResolver
	DependantsOf(string) []string
}

type defaultSnake struct {
	resolvers  map[string]UntypedResolver
	dependants map[string][]string
}

func (me *defaultSnake) ResolverNames() []string {
	names := make([]string, 0)
	for k := range me.resolvers {
		names = append(names, k)
	}
	return names
}

func (me *defaultSnake) Resolve(name string) UntypedResolver {
	return me.resolvers[name]
}

type SnakeImplementationTyped[X any] interface {
	Decorate(context.Context, TypedResolver[X], Snake, []Input, []Middleware) error
	SnakeImplementation
}

type SnakeImplementation interface {
	ManagedResolvers(context.Context) []UntypedResolver
	OnSnakeInit(context.Context, Snake) error
	ResolveEnum(string, []string) (string, error)
	ProvideContextResolver() UntypedResolver
}

func NewSnake[M NamedMethod](ctx context.Context, impl SnakeImplementationTyped[M], res ...UntypedResolver) (Snake, error) {
	return NewSnakeWithOpts(ctx, impl, &NewSnakeOpts[M]{
		Resolvers: res,
	})
}

func snakeManagedResolvers() []UntypedResolver {
	return []UntypedResolver{
		NewNoopMethod[Chan](),
		NewNoopMethod[io.Writer](),
		NewNoopMethod[io.Reader](),
		NewNoopMethod[Stdin](),
		NewNoopMethod[Stdout](),
		NewNoopMethod[Stderr](),
	}
}

func NewSnakeWithOpts[M Method](ctx context.Context, impl SnakeImplementationTyped[M], opts *NewSnakeOpts[M]) (Snake, error) {
	var err error

	snk := &defaultSnake{
		resolvers:  make(map[string]UntypedResolver),
		dependants: make(map[string][]string),
	}

	enums := make([]Enum, 0)

	named := make(map[string]TypedResolver[M])

	inputResolvers := make([]UntypedResolver, 0)

	if opts.Resolvers != nil {
		inputResolvers = append(inputResolvers, opts.Resolvers...)
	}

	inputResolvers = append(inputResolvers, NewResolvedResolver[EnumResolverFunc](impl.ResolveEnum))

	con := impl.ProvideContextResolver()
	if con != nil {
		inputResolvers = append(inputResolvers, con)
	}

	inputResolvers = append(inputResolvers, impl.ManagedResolvers(ctx)...)

	inputResolvers = append(inputResolvers, snakeManagedResolvers()...)

	for _, runner := range opts.Commands {
		named[runner.Name()] = runner
	}

	for _, runner := range inputResolvers {

		if nmd, ok := runner.(TypedResolver[M]); ok {
			named[nmd.Name()] = nmd
			continue
		}

		retrn := ListOfReturns(runner)

		// every return value marks this runner as the resolver for that type
		for _, r := range retrn {
			if r.Kind().String() == "error" {
				continue
			}
			snk.resolvers[reflectTypeString(r)] = runner
		}

		// enum options are also resolvers so they are passed here
		if mp, ok := runner.(Enum); ok {
			resolver := opts.OverrideEnumResolver
			if resolver == nil {
				resolver = impl.ResolveEnum
			}
			err := mp.ApplyResolver(resolver)
			if err != nil {
				return nil, err
			}
			enums = append(enums, mp)
		}

	}

	for name, runner := range named {
		snk.resolvers[name] = runner

		inpts, err := DependancyInputs(name, snk.Resolve, enums...)
		if err != nil {
			return nil, err
		}

		mw := make([]Middleware, 0)

		if mwd, ok := runner.(MiddlewareProvider); ok {
			mw = append(mw, mwd.Middlewares()...)

			for _, m := range mwd.Middlewares() {

				mwin, err := InputsFor(NewMiddlewareResolver(m), enums...)
				if err != nil {
					return nil, err
				}

				inpts = append(inpts, mwin...)
			}
		}

		err = impl.Decorate(ctx, runner, snk, inpts, mw)
		if err != nil {
			return nil, err
		}

	}

	for name := range snk.resolvers {

		deps, err := DependanciesOf(name, snk.Resolve)
		if err != nil {
			return nil, terrors.Wrapf(err, "failed to find dependancies of %q", name)
		}

		for _, dep := range deps {
			methd := MethodName(snk.Resolve(dep))

			if _, ok := snk.dependants[methd]; !ok {
				snk.dependants[methd] = make([]string, 0)
			}

			// fmt.Println("adding dependant", name, "to", methd)
			snk.dependants[methd] = append(snk.dependants[methd], name)
		}
	}

	err = impl.OnSnakeInit(ctx, snk)
	if err != nil {
		return nil, err
	}

	return snk, nil

}

func buildMiddlewareName(name string, m Middleware) string {
	return name + "_" + reflectTypeString(reflect.TypeOf(m))
}

func (me *defaultSnake) Enums() []Enum {
	enums := make([]Enum, 0)
	for _, name := range me.resolvers {
		if mp, ok := name.(Enum); ok {
			enums = append(enums, mp)
		}
	}
	return enums
}

func (me *defaultSnake) Resolvers() []UntypedResolver {
	abc := make([]UntypedResolver, 0)
	for _, name := range me.resolvers {
		abc = append(abc, name)
	}
	return abc
}

func (me *defaultSnake) DependantsOf(name string) []string {
	return me.dependants[name]
}

func Commands[M any](cmds ...TypedResolver[M]) []TypedResolver[M] {
	return cmds
}

type backwardsResolver struct {
	LegacyNamedRunner
}

type LegacyNamedRunner interface {
	Runner
	Named
}

func (me *backwardsResolver) RegisterRunFunc() RunFunc {
	return me.LegacyNamedRunner
}

func (me *backwardsResolver) Name() string {
	return me.LegacyNamedRunner.Name()
}

func (me *backwardsResolver) Description() string {
	return me.LegacyNamedRunner.Description()
}

func Command[I SnakeImplementationTyped[M], M Method, Rnr LegacyNamedRunner](runner func() Rnr, impl I, cmd M) TypedResolver[M] {
	return NewInlineRunner(cmd, &backwardsResolver{runner()})
}

func Resolvers(args ...UntypedResolver) []UntypedResolver {
	return args
}
func Resolver(runner func() Runner) UntypedResolver {
	return runner()
}

func ResolverV2(runner RegisterableRunFunc) UntypedResolver {
	return runner.RegisterRunFunc()
}
