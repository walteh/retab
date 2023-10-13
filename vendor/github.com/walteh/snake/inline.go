package snake

import (
	"reflect"

	"github.com/spf13/pflag"
)

var _ Flagged = (*inlineResolver[any])(nil)

type inlineResolver[I any] struct {
	flagFunc func(*pflag.FlagSet)
	runFunc  func() (I, error)
}

func (me *inlineResolver[I]) Flags(flgs *pflag.FlagSet) {
	me.flagFunc(flgs)
}

func (me *inlineResolver[I]) Run() (I, error) {
	return me.runFunc()
}

func NewArgInlineFunc[I any](flagFunc func(*pflag.FlagSet), runFunc func() (I, error)) Flagged {
	return &inlineResolver[I]{flagFunc: flagFunc, runFunc: runFunc}
}

func (me *inlineResolver[I]) AsArgumentMethod(name string) Method {
	return &method{
		name:             name,
		flags:            me.flagFunc,
		responseStrategy: handleArgumentResponse[I],
		method: reflect.ValueOf(func() (I, error) {
			return me.Run()
		}),
		validationStrategy: validateArgumentResponse[I],
	}
}

// func NewInlineFuncSimple[I any](runFunc func() (I, error)) Flagged {
// 	return &inlineResolver[I]{flagFunc: func(*pflag.FlagSet) {}, runFunc: runFunc}
// }

// func NewInlineSimple[I any](value I) Flagged {
// 	return &inlineResolver[I]{flagFunc: func(*pflag.FlagSet) {}, runFunc: func() (I, error) { return value, nil }}
// }
