package snake

import (
	"reflect"
)

var _ MiddlewareProvider = &inlineResolver[any]{}

type inlineResolver[M any] struct {
	Runner
	trickref    M
	middlewares []Middleware
	name        string
	description string
}

func (me *inlineResolver[M]) RunFunc() reflect.Value {
	return me.Runner.RunFunc()
}

func (me *inlineResolver[M]) Ref() Method {
	return me.Runner.Ref()
}

func (me *inlineResolver[M]) TypedRef() M {
	return me.trickref
}

func (me *inlineResolver[M]) IsResolver() {
}

func (me *inlineResolver[M]) Middlewares() []Middleware {
	return me.middlewares
}

func (me *inlineResolver[M]) WithMiddleware(mw ...Middleware) TypedResolver[M] {
	me.middlewares = append(me.middlewares, mw...)
	return me
}

func (me *inlineResolver[M]) WithName(name string) TypedResolver[M] {
	me.name = name
	return me
}

func (me *inlineResolver[M]) WithTypedRef(m M) TypedResolver[M] {
	me.trickref = m
	return me
}

func (me *inlineResolver[M]) WithDescription(description string) TypedResolver[M] {
	me.description = description
	return me
}

func (me *inlineResolver[M]) WithRunner(m func() Runner) TypedResolver[M] {
	me.Runner = m()
	return me
}

func (me *inlineResolver[M]) Name() string {
	return me.name
}

func (me *inlineResolver[M]) Description() string {
	return me.description
}

type RegisterableRunFunc interface {
	RegisterRunFunc() RunFunc
}

func NewInlineNamedRunner[M any](typed M, nmd RegisterableRunFunc, name, desc string) TypedResolver[M] {
	return &inlineResolver[M]{
		Runner:      nmd.RegisterRunFunc(),
		trickref:    typed,
		middlewares: []Middleware{},
		name:        name,
		description: desc,
	}
}

func NewInlineRunner[M any](typed M, nmd RegisterableRunFunc) TypedResolver[M] {
	if nmdz, ok := nmd.(Named); ok {
		return NewInlineNamedRunner[M](typed, nmd, nmdz.Name(), nmdz.Description())
	}
	panic("not a named runner")
}
