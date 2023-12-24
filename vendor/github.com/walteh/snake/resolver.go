package snake

import (
	"reflect"

	"github.com/walteh/terrors"
)

var (
	_ UntypedResolver       = (*simpleResolver[Method])(nil)
	_ TypedResolver[Method] = (*simpleResolver[Method])(nil)
	_ MiddlewareProvider    = (*simpleResolver[Method])(nil)
)

type TypedResolver[M Method] interface {
	UntypedResolver
	TypedRef() M
	WithRunner(func() Runner) TypedResolver[M]
	WithMiddleware(...Middleware) TypedResolver[M]
	WithName(string) TypedResolver[M]
	WithDescription(string) TypedResolver[M]
	WithTypedRef(M) TypedResolver[M]
	Name() string
	Description() string
}

type UntypedResolver interface {
	RunFunc() reflect.Value
	Ref() Method
	IsResolver()
}

type MethodProvider interface {
	Method() reflect.Value
}

type simpleResolver[M Method] struct {
	runfunc     reflect.Value
	ref         Method
	typedRef    M
	middlewares []Middleware
	name        string
	description string
}

func NewResolvedResolver[M Method](strc M) TypedResolver[M] {
	return &simpleResolver[M]{
		runfunc: reflect.ValueOf(func() (M, error) {
			return strc, nil
		}),
		ref:      strc,
		typedRef: strc,
	}
}

func (me *simpleResolver[M]) WithRunner(m func() Runner) TypedResolver[M] {
	rnr := m()
	me.ref = rnr.Ref()
	me.runfunc = rnr.RunFunc()
	return me
}

func (me *simpleResolver[M]) RunFunc() reflect.Value {
	return me.runfunc
}

func (me *simpleResolver[M]) Ref() Method {
	return me.ref
}

func (me *simpleResolver[M]) TypedRef() M {
	return me.typedRef
}

func (me *simpleResolver[M]) Name() string {
	return me.name
}

func (me *simpleResolver[M]) Description() string {
	return me.description
}

func (me *simpleResolver[M]) WithMiddleware(mw ...Middleware) TypedResolver[M] {
	me.middlewares = append(me.middlewares, mw...)
	return me
}

func (me *simpleResolver[M]) WithTypedRef(m M) TypedResolver[M] {
	me.typedRef = m
	return me
}

func (me *simpleResolver[M]) WithName(name string) TypedResolver[M] {
	me.name = name
	return me
}

func (me *simpleResolver[M]) WithDescription(desc string) TypedResolver[M] {
	me.description = desc
	return me
}

func (me *simpleResolver[M]) Middlewares() []Middleware {
	return me.middlewares
}

func (me *simpleResolver[M]) IsResolver() {}

func MustGetTypedResolver[M Method](inter M) TypedResolver[M] {

	m, err := getTypedResolver(inter)
	if err != nil {
		panic(err)
	}
	return m
}

func MustGetResolverFor[M any](inter Method) UntypedResolver {
	return mustGetResolverForRaw(inter, (*M)(nil))
}

func MustGetResolverFor2[M1, M2 any](inter Method) UntypedResolver {
	return mustGetResolverForRaw(inter, (*M1)(nil), (*M2)(nil))
}

func MustGetResolverFor3[M1, M2, M3 any](inter Method) UntypedResolver {
	return mustGetResolverForRaw(inter, (*M1)(nil), (*M2)(nil), (*M3)(nil))
}

func mustGetResolverForRaw(inter any, args ...any) UntypedResolver {
	run, err := getTypedResolver(inter)
	if err != nil {
		panic(err)
	}

	resvf := IsResolverFor(run)

	for _, arg := range args {
		argptr := reflect.TypeOf(arg).Elem()
		if yes, ok := resvf[argptr.String()]; !ok || !yes {
			panic(terrors.Errorf("%q is not a resolver for %q", reflect.TypeOf(inter).String(), argptr.String()))
		}
	}

	return run
}
func getTypedResolver[M Method](inter M) (TypedResolver[M], error) {

	if m, ok := any(inter).(Runner); ok {
		return m.(TypedResolver[M]), nil
	}

	prov, ok := any(inter).(MethodProvider)
	if ok {
		return &simpleResolver[M]{
			runfunc:  prov.Method(),
			ref:      inter,
			typedRef: inter,
		}, nil
	}

	value := reflect.ValueOf(inter)

	method := value.MethodByName("Run")
	if !method.IsValid() {
		if value.CanAddr() {
			method = value.Addr().MethodByName("Run")
		}
	}

	if !method.IsValid() {
		return nil, terrors.Errorf("missing Run method on %q", value.Type())
	}

	sr := &simpleResolver[M]{
		runfunc:  method,
		ref:      inter,
		typedRef: inter,
	}

	if name, ok := any(inter).(NamedMethod); ok {
		sr.name = name.Name()
		sr.description = name.Description()
	}

	return sr, nil
}

func ListOfArgs(m UntypedResolver) []reflect.Type {
	var args []reflect.Type
	typ := m.RunFunc().Type()
	for i := 0; i < typ.NumIn(); i++ {
		args = append(args, typ.In(i))
	}

	return args
}

func ListOfReturns(m UntypedResolver) []reflect.Type {
	var args []reflect.Type
	typ := m.RunFunc().Type()
	for i := 0; i < typ.NumOut(); i++ {
		args = append(args, typ.Out(i))
	}
	return args
}

func MenthodIsShared(run UntypedResolver) bool {
	rets := ListOfReturns(run)
	// right now this logic relys on the fact that commands only return one value (the error)
	// and shared methods return two or more (the error and the values)
	if len(rets) == 1 ||
		// this is the logic to support the new Output type
		(len(rets) == 2 && rets[0].String() == reflect.TypeOf((*Output)(nil)).Elem().String()) {
		return false
	} else {
		return true
	}
}

func IsResolverFor(m UntypedResolver) map[string]bool {
	resp := make(map[string]bool, 0)
	for _, f := range ListOfReturns(m) {
		if f.String() == "error" {
			continue
		}
		resp[f.String()] = true
	}
	return resp
}

func FieldByName(me UntypedResolver, name string) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(me.Ref()).Elem()).FieldByName(name)
}

func CallMethod(me UntypedResolver, args []reflect.Value) []reflect.Value {
	return me.RunFunc().Call(args)
}

func StructFields(me UntypedResolver) []reflect.StructField {
	typ := reflect.TypeOf(me.Ref())
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return []reflect.StructField{}
	}
	vis := reflect.VisibleFields(typ)
	return vis
}
