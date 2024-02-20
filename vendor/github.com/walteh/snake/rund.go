package snake

import "reflect"

type RunFunc Runner

// A runner is different from a resolver in that it does not need to have any extra validation
// the only Runner that i exists is rund[X], which is generically validated
type Runner interface {
	isRunner()
	UntypedResolver
}

type TypedRunner[X any] interface {
	Runner
	TypedRef() X
}
type TypedNamedRunner[X any] interface {
	NamedRunner
	TypedRunner[X]
}

type NamedRunner interface {
	Runner
}

type Named interface {
	Name() string
	Description() string
}

var _ Runner = (*rund[Method])(nil)

type rund[X any] struct {
	internal X
}

type namedrund[X NamedMethod] struct {
	*rund[X]
}

// func (r *namedrund[X]) Name() string {
// 	return r.internal.Name()
// }

// func (r *namedrund[X]) Description() string {
// 	return r.internal.Description()
// }

func (r *rund[X]) IsResolver() {}

func (r *rund[X]) isRunner() {}

func (r *rund[X]) RunFunc() reflect.Value {
	return reflect.ValueOf(r.internal).MethodByName("Run")
}

func (r *rund[X]) Ref() Method {
	return any(r.internal).(Method)
}

func (r *rund[X]) TypedRef() X {
	return r.internal
}

// type Runnable[M any] interface {
// 	NamedMethod
// 	RunMethod() TypedRunner[M]
// }
