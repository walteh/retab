package snake

import (
	"reflect"

	"github.com/walteh/terrors"
)

type Method interface {
}

type NamedMethod interface {
	Name() string
	Description() string
}

type FMap func(string) UntypedResolver

func DependanciesOf(str string, m FMap) ([]string, error) {
	if ok := m(str); !reflect.ValueOf(ok).IsValid() || reflect.ValueOf(ok).IsNil() {
		return nil, terrors.Errorf("missing resolver for %q", str)
	}

	mapa, err := FindBrothers(str, m, ListOfArgs)
	if err != nil {
		return nil, err
	}

	return mapa, nil
}

func DependantsOf(str string, m FMap) ([]string, error) {
	mapa, err := FindBrothers(str, m, ListOfReturns)
	if err != nil {
		return nil, err
	}

	return mapa, nil
}

func EndOfChain() reflect.Value {
	return reflect.ValueOf("end_of_chain")
}

func EndOfChainPtr() *reflect.Value {
	v := EndOfChain()
	return &v
}

type ListFunc func(UntypedResolver) []reflect.Type

func FindBrothers(str string, me FMap, listFunc ListFunc) ([]string, error) {
	raw, err := findBrothersRaw(str, me, nil, listFunc)
	if err != nil {
		return nil, err
	}
	resp := make([]string, 0)
	for k := range raw {
		resp = append(resp, k)
	}
	return resp, nil
}

func findBrothersRaw(str string, fmap FMap, rmap map[string]bool, listFunc ListFunc) (map[string]bool, error) {
	var err error
	if rmap == nil {
		rmap = make(map[string]bool)
	}

	validated := fmap(str)

	if validated == nil {
		return nil, terrors.Errorf("missing resolver for %q", str)
	}

	if rmap[str] {
		return rmap, nil
	}

	rmap[str] = true

	for _, f := range listFunc(validated) {
		rmap, err = findBrothersRaw(f.String(), fmap, rmap, listFunc)
		if err != nil {
			return nil, err
		}
	}

	return rmap, nil
}

func FindArguments(str string, fmap FMap) ([]reflect.Value, error) {
	raw, err := findArgumentsRaw(str, fmap, nil)
	if err != nil {
		return nil, err
	}
	resp := make([]reflect.Value, 0)
	for _, v := range raw.bindings {
		resp = append(resp, *v)
	}
	return resp, nil
}

func reflectTypeString(typ reflect.Type) string {
	return typ.String()
}

func findArgumentsRaw(str string, fmap FMap, wrk *Binder) (*Binder, error) {
	validated := fmap(str)
	var err error
	if validated == nil {
		return nil, terrors.Errorf("missing resolver for %q", str)
	}

	if wrk == nil {
		wrk = NewBinder()
	}

	if _, ok := wrk.bindings[str]; ok {
		return wrk, nil
	}

	tmp := make([]reflect.Value, 0)
	for _, f := range ListOfArgs(validated) {
		name := reflectTypeString(f)
		wrk, err = findArgumentsRaw(name, fmap, wrk)
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, *wrk.bindings[name])
	}

	out := CallMethod(validated, tmp)

	if !MenthodIsShared(validated) {
		// only commands can have one response value, which is always an error
		// so here we know we can name it str
		// otherwise we would be naming it "error"
		//
		if len(out) == 1 {
			wrk.Bind(str, &out[0])
			if out[0].Interface() != nil {
				return wrk, out[0].Interface().(error)
			} else {
				// we want to get out right away as we know this is an error (only one return value)
				return wrk, nil
			}
		}

		resp := out[0]

		if resp.IsNil() {
			resp = reflect.ValueOf(&NilOutput{})
		}

		wrk.Bind(str, &resp)

	} else {
		for _, v := range out {
			in := v
			strd := v.Type().String()
			if strd != "error" {
				wrk.Bind(strd, &in)
			} else {
				if in.Interface() != nil {
					return wrk, in.Interface().(error)
				}
			}
		}
	}

	return wrk, nil

}
