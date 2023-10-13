package snake

import (
	"reflect"

	"github.com/go-faster/errors"
	"github.com/spf13/pflag"
)

var (
	ErrMissingResolver = errors.New("missing resolver")
)

type HasRunArgs interface{ RunArgs() []reflect.Type }

type IsRunnable interface {
	HasRunArgs
	Run() reflect.Value
	HandleResponse([]reflect.Value) (*reflect.Value, error)
}

type FMap[G any] func(string) G

func (me *Ctx) FlagsFor(str Method) (*pflag.FlagSet, error) {
	return me.FlagsForString(str.Name())
}

func (me *Ctx) FlagsForString(str string) (*pflag.FlagSet, error) {
	if _, ok := me.resolvers[str]; !ok {
		return nil, errors.Wrapf(ErrMissingResolver, "missing resolver for %q", str)
	}

	mapa, err := findBrothers(str, func(s string) HasRunArgs {
		return me.resolvers[s]
	})
	if err != nil {
		return nil, err
	}

	flgs := &pflag.FlagSet{}

	for _, f := range mapa {
		if z, ok := me.resolvers[f]; !ok {
			return nil, errors.Wrapf(ErrMissingResolver, "missing resolver for %q", f)
		} else {
			z.Flags(flgs)
		}
	}

	return flgs, nil
}

func (me *Ctx) Run(str Method) error {
	return me.RunString(str.Name())
}

func (me *Ctx) RunString(str string) error {
	args, err := findArgumentsRaw(str, func(s string) IsRunnable {
		return me.resolvers[s]
	}, nil)
	if err != nil {
		return err
	}

	if resp, ok := args[str]; !ok {
		return errors.Wrapf(ErrMissingResolver, "missing resolver for %q", str)
	} else {
		if resp == end_of_chain_ptr {
			return nil
		} else {
			return errors.Errorf("expected end of chain, got %v", resp)
		}
	}

}

func findBrothers(str string, me FMap[HasRunArgs]) ([]string, error) {
	raw, err := findBrothersRaw(str, me, nil)
	if err != nil {
		return nil, err
	}
	resp := make([]string, 0)
	for k := range raw {
		resp = append(resp, k)
	}
	return resp, nil
}

var Defaults = []string{"context.Context", "*cobra.Command"}

func findBrothersRaw(str string, fmap FMap[HasRunArgs], rmap map[string]bool) (map[string]bool, error) {
	var err error
	if rmap == nil {
		rmap = make(map[string]bool)
		rmap["context.Context"] = true
		// rmap["*cobra.Command"] = true
	}

	var curr HasRunArgs

	if ok := fmap(str); ok == nil {
		return nil, errors.Wrapf(ErrMissingResolver, "missing resolver for %q", str)
	} else {
		curr = ok
	}

	if rmap[str] {
		return rmap, nil
	}

	rmap[str] = true

	for _, f := range curr.RunArgs() {
		rmap, err = findBrothersRaw(f.String(), fmap, rmap)
		if err != nil {
			return nil, err
		}
	}

	return rmap, nil
}

func findArguments(str string, fmap FMap[IsRunnable]) ([]reflect.Value, error) {
	raw, err := findArgumentsRaw(str, fmap, nil)
	if err != nil {
		return nil, err
	}
	resp := make([]reflect.Value, 0)
	for _, v := range raw {
		resp = append(resp, *v)
	}
	return resp, nil
}

func valueToIsRunnable(v reflect.Value) IsRunnable {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Interface().(IsRunnable)
}

func runResolvingArguments(str string, fmap FMap[IsRunnable], bmap map[string]*reflect.Value) error {

	args, err := findArgumentsRaw(str, fmap, bmap)
	if err != nil {
		return err
	}

	if resp, ok := args[str]; !ok {
		return errors.Wrapf(ErrMissingResolver, "missing resolver for %q", str)
	} else {
		if resp == end_of_chain_ptr {
			return nil
		} else {
			return errors.Errorf("expected end of chain, got %v", resp)
		}
	}

}

func reflectTypeString(typ reflect.Type) string {
	return typ.String()
}

func findArgumentsRaw(str string, fmap FMap[IsRunnable], wrk map[string]*reflect.Value) (map[string]*reflect.Value, error) {
	var curr IsRunnable
	var err error
	if ok := fmap(str); ok == nil {
		return nil, errors.Wrapf(ErrMissingResolver, "missing resolver for %q", str)
	} else {
		curr = ok
	}

	if wrk == nil {
		wrk = make(map[string]*reflect.Value)
	}

	if _, ok := wrk[str]; ok {
		return wrk, nil
	}

	tmp := make([]reflect.Value, 0)
	for _, f := range curr.RunArgs() {
		name := reflectTypeString(f)
		wrk, err = findArgumentsRaw(name, fmap, wrk)
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, *wrk[name])
	}

	resp := curr.Run().Call(tmp)
	out, err := curr.HandleResponse(resp)
	if err != nil {
		return nil, err
	}

	wrk[str] = out

	return wrk, nil

}
