package snake

import (
	"context"
	"reflect"
	"time"
)

type Refreshable interface {
	RefreshInterval() time.Duration
}

type MiddlewareFunc func(ctx context.Context) error

type Middleware interface {
	Method
	Wrap(MiddlewareFunc) MiddlewareFunc
}

type MiddlewareProvider interface {
	Middlewares() []Middleware
}

type middlewareResolver struct {
	mw Middleware
}

func (*middlewareResolver) IsResolver() {}

func NewMiddlewareResolver(mw Middleware) UntypedResolver {
	return &middlewareResolver{
		mw: mw,
	}
}

func (me *middlewareResolver) Name() string {
	return "middleware"
}

func (me *middlewareResolver) Description() string {
	return "middleware"
}

func (me *middlewareResolver) Ref() Method {
	return me.mw
}

func (me *middlewareResolver) RunFunc() reflect.Value {
	// this *struct{} makes the resolver "Shared"
	return reflect.ValueOf(func(ctx context.Context) (*struct{}, error) {
		return nil, nil
	})
}
