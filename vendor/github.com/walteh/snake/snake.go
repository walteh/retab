package snake

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/spf13/cobra"

	"github.com/go-faster/errors"
)

type Snakeable interface {
	PreRun(ctx context.Context, args []string) (context.Context, error)
	Register(context.Context) (*cobra.Command, error) // need a way for us to be able to have the full command like we do when the kubectl command is importable
}

var (
	ErrMissingBinding   = errors.New("snake.ErrMissingBinding")
	ErrMissingRun       = errors.New("snake.ErrMissingRun")
	ErrInvalidRun       = errors.New("snake.ErrInvalidRun")
	ErrInvalidArguments = errors.New("snake.ErrInvalidArguments")
	ErrInvalidResolver  = errors.New("snake.ErrInvalidResolver")
)

func NewRootCommand(ctx context.Context, snk Snakeable) (context.Context, error) {

	cmd, err := snk.Register(ctx)
	if err != nil {
		panic(err)
	}

	nc := &NamedCommand{
		cmd:        cmd,
		method:     reflect.ValueOf(func() {}),
		methodType: reflect.TypeOf(func() {}),
		ptr:        nil,
	}

	ctx = SetNamedCommand(ctx, RootCommandName, nc)

	ctx = SetRootCommand(ctx, nc)

	cmd.SilenceErrors = true

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		zctx := cmd.Context()

		zctx = SetRootCommand(zctx, nc)
		zctx = SetActiveCommand(zctx, RootCommandName)
		defer func() {
			zctx = ClearActiveCommand(zctx)
		}()

		if err := cmd.ParseFlags(args); err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}

		zctx, err := snk.PreRun(zctx, args)
		if err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}

		cmd.SetContext(zctx)

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		zctx := cmd.Context()
		zctx = SetRootCommand(zctx, nc)
		zctx = SetActiveCommand(zctx, RootCommandName)
		defer func() {
			zctx = ClearActiveCommand(zctx)
		}()

		zctx, err := snk.PreRun(zctx, args)
		if err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}

		cmd.SetContext(zctx)

		return nil
	}

	return ctx, nil
}

type NamedCommand struct {
	cmd        *cobra.Command
	method     reflect.Value
	methodType reflect.Type
	ptr        any
}

func NewCommand(ctx context.Context, name string, snk Snakeable) (context.Context, error) {

	cmd, err := snk.Register(ctx)
	if err != nil {
		return nil, err
	}

	method := getRunMethod(snk)

	tpe, err := validateRunMethod(snk, method)
	if err != nil {
		return nil, err
	}

	rootcmd := GetRootCommand(ctx)
	if rootcmd == nil {
		return nil, fmt.Errorf("snake.NewCommand: no root command found in context")
	}

	getRootCommandCtx := func() context.Context {
		if c, ok := rootcmd.ptr.(context.Context); ok {
			return c
		}
		return ctx
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		zctx := cmd.Context()

		zctx = SetRootCommand(zctx, rootcmd)
		zctx = SetActiveCommand(zctx, name)
		defer func() {
			zctx = ClearActiveCommand(zctx)
		}()

		if err := cmd.ParseFlags(args); err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}

		zctx, err := snk.PreRun(zctx, args)
		if err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}

		cmd.SetContext(zctx)

		return nil
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		zctx := cmd.Context()
		zctx = SetRootCommand(zctx, rootcmd)
		zctx = SetActiveCommand(zctx, name)
		defer func() {
			zctx = ClearActiveCommand(zctx)
		}()

		dctx, err := ResolveBindingsFromProvider(getRootCommandCtx(), method)
		if err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}

		zctx = mergeBindingKeepingFirst(zctx, dctx)

		cmd.SetContext(zctx)

		if err := callRunMethod(cmd, method, tpe); err != nil {
			return HandleErrorByPrintingToConsole(cmd, err)
		}
		return nil
	}

	if name != "" {
		cmd.Use = name
	}

	ctx = SetNamedCommand(ctx, name, &NamedCommand{
		cmd:        cmd,
		method:     method,
		methodType: tpe,
	})

	return ctx, nil
}

func Assemble(ctx context.Context) *cobra.Command {
	rootcmd := GetRootCommand(ctx)
	if rootcmd == nil {
		return nil
	}

	named := GetAllNamedCommands(ctx)
	if named == nil {
		return nil
	}

	flagb := getFlagBindings(ctx)

	for name, cmd := range named {
		if name == RootCommandName {
			continue
		}

		for _, arg := range listOfArgs(cmd.methodType) {
			if flagb[arg.String()] == nil {
				continue
			}
			flagb[arg.String()](cmd.cmd.Flags())
		}

		rootcmd.cmd.AddCommand(cmd.cmd)
	}

	// we keep the context so we can use it to resolve bindings when a command is run
	rootcmd.ptr = ctx

	return rootcmd.cmd
}

func MustNewCommand(ctx context.Context, name string, snk Snakeable) context.Context {
	ctx, err := NewCommand(ctx, name, snk)
	if err != nil {
		panic(err)
	}
	return ctx
}

func GetAlreadyBound[I any](ctx context.Context) (I, bool) {
	// allocate a new I
	fake := *new(I)

	b, ok := ctx.Value(&bindingsKeyT{}).(bindings)
	if !ok {
		return fake, false
	}

	ft := reflect.TypeOf(fake)
	if ft == nil {
		ft = reflect.TypeOf((*I)(nil)).Elem()
	}
	br, ok := b[ft.String()]
	if !ok {
		return fake, false
	}
	brs, err := br()
	if err != nil {
		return fake, false
	}

	brsl, ok := brs.Interface().(I)
	if !ok {
		return fake, false
	}

	return brsl, true
}

func WithRootCommand(ctx context.Context, x func(*cobra.Command) error) error {
	root := GetRootCommand(ctx)
	if root == nil {
		return fmt.Errorf("snake.WithRootCommand: no root command found in context")
	}
	return x(root.cmd)
}

func WithNamedCommand(ctx context.Context, name string, x func(*cobra.Command) error) error {
	cmd := GetNamedCommand(ctx, name)
	if cmd == nil {
		return fmt.Errorf("snake.WithNamedCommand: no named command found in context")
	}
	return x(cmd.cmd)
}

func WithActiveCommand(ctx context.Context, x func(*cobra.Command) error) error {
	cmd := GetActiveNamedCommand(ctx)
	if cmd == nil {
		return fmt.Errorf("snake.WithActiveCommand: no active command found in context")
	}
	return x(cmd.cmd)
}

const RootCommandName = "______root_____________"

type rootKeyT struct {
}

var rootKey = rootKeyT{}

func SetRootCommand(ctx context.Context, cmd *NamedCommand) context.Context {
	return context.WithValue(ctx, rootKey, cmd)
}

func GetRootCommand(ctx context.Context) *NamedCommand {
	p, ok := ctx.Value(rootKey).(*NamedCommand)
	if ok {
		return p
	}
	return nil
}

type namedCommandKeyT struct {
}

var namedCommandKey = namedCommandKeyT{}

type namedCommandMap map[string]*NamedCommand

var namedCommandMutex = sync.RWMutex{}

func SetNamedCommand(ctx context.Context, name string, cmd *NamedCommand) context.Context {

	ncm, ok := ctx.Value(namedCommandKey).(namedCommandMap)
	if !ok {
		ncm = make(namedCommandMap)
	}
	namedCommandMutex.Lock()
	ncm[name] = cmd
	namedCommandMutex.Unlock()

	return context.WithValue(ctx, namedCommandKey, ncm)
}

func GetNamedCommand(ctx context.Context, name string) *NamedCommand {
	p, ok := ctx.Value(namedCommandKey).(namedCommandMap)
	if ok {
		namedCommandMutex.RLock()
		defer namedCommandMutex.RUnlock()
		return p[name]
	}
	return nil
}

func GetAllNamedCommands(ctx context.Context) namedCommandMap {
	p, ok := ctx.Value(namedCommandKey).(namedCommandMap)
	if ok {
		namedCommandMutex.RLock()
		defer namedCommandMutex.RUnlock()
		return p
	}
	return nil
}

type activeCommandKeyT struct {
}

var activeCommandKey = activeCommandKeyT{}

func SetActiveCommand(ctx context.Context, str string) context.Context {
	return context.WithValue(ctx, activeCommandKey, str)
}

func GetActiveCommand(ctx context.Context) string {
	p, ok := ctx.Value(activeCommandKey).(string)
	if ok {
		return p
	}
	return ""
}

func ClearActiveCommand(ctx context.Context) context.Context {
	return context.WithValue(ctx, activeCommandKey, "")
}

func GetActiveNamedCommand(ctx context.Context) *NamedCommand {
	return GetNamedCommand(ctx, GetActiveCommand(ctx))
}

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
