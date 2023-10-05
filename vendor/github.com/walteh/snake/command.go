package snake

import (
	"reflect"
	"sync"

	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
)

type bindingmap map[string]*reflect.Value

type Ctx struct {
	bindings  map[string]*reflect.Value
	resolvers map[string]Method
	cmds      map[string]*cobra.Command

	runlock sync.Mutex
}

var (
	ErrMissingResolver = errors.New("missing resolver")
)

func setBindingWithLock[T any](con *Ctx, val T) func() {
	con.runlock.Lock()
	ptr := reflect.ValueOf(val)
	typ := reflect.TypeOf((*T)(nil)).Elem()
	con.bindings[typ.String()] = &ptr
	return func() {
		delete(con.bindings, typ.String())
		con.runlock.Unlock()
	}
}

// func (me *Ctx) getResolvedRunArgs(cmd Method) ([]reflect.Value, error) {
// 	args := cmd.RunArgs()
// 	rargs := make([]reflect.Value, len(args))
// 	for i, arg := range args {
// 		// first check if we have a binding for this type
// 		if bnd, ok := me.bindings[arg.String()]; ok {
// 			rargs[i] = *bnd
// 		} else {
// 			// if not, check if we have a resolver for this type
// 			if resl, ok := me.resolvers[arg.String()]; ok {
// 				// if we do, run it, and store the result as a binding
// 				// its okay to recurse here, because we are saving the result as a binding
// 				bnd, err := me.run(resl)
// 				if err != nil {
// 					return nil, err
// 				}
// 				rargs[i] = bnd
// 				me.bindings[arg.String()] = &bnd
// 			} else {
// 				return nil, errors.Wrapf(ErrMissingResolver, "missing resolver for type %q", arg.String())
// 			}
// 		}
// 	}
// 	return rargs, nil
// }

// func (me *Ctx) run(cmd Method) (*reflect.Value, error) {

// 	args, err := me.getResolvedRunArgs(cmd)
// 	if err != nil {
// 		return nil, err
// 	}

// 	out := cmd.Run().Call(args)

// 	return cmd.HandleResponse(out)
// }

var (
	ErrInvalidMethodSignature = errors.New("invalid method signature")
)

var end_of_chain = reflect.ValueOf("end_of_chain")
var end_of_chain_ptr = &end_of_chain

var prefix_command = "snake_command_"
var prefix_argument = "snake_argument_"
