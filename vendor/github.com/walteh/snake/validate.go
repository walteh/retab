package snake

import (
	"reflect"

	"github.com/walteh/terrors"
)

type Strategy interface {
	ValidateResponseTypes([]reflect.Type) error
	// HandleResponse([]reflect.Value) ([]*reflect.Value, error)
}

type CommandStrategy struct {
}

func (me *CommandStrategy) ValidateResponseTypes(out []reflect.Type) error {

	if len(out) != 1 {
		return terrors.Errorf("invalid return signature, expected 1, got %d", len(out))
	}

	if !out[0].Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return terrors.Errorf("invalid return type %q, expected %q", out[0].String(), reflect.TypeOf((*error)(nil)).Elem().String())
	}

	return nil
}

// func (me *CommandStrategy) HandleResponse(out []reflect.Value) ([]*reflect.Value, error) {

// 	eoc := EndOfChain()

// 	resp := []*reflect.Value{&eoc}

// 	if !out[0].IsNil() {
// 		return resp, out[0].Interface().(error)
// 	}

// 	return resp, nil
// }

func NewCommandStrategy() *CommandStrategy {
	return &CommandStrategy{}
}

type ArgumentStrategy struct {
	args []any
}

// func (me *ArgumentStrategy) HandleResponse(out []reflect.Value) ([]*reflect.Value, error) {

// 	res := make([]*reflect.Value, len(me.args))

// 	if !out[len(out)-1].IsNil() {
// 		// need to fix this TODO
// 		return nil, out[len(out)-1].Interface().(error)
// 	}

// 	for i, v := range me.args {
// 		if out[i].Type() != reflect.TypeOf(v).Elem() {
// 			return nil, errors.Wrapf(ErrInvalidMethodSignature, "invalid return type %q, expected %q", out[i].String(), reflect.TypeOf(v).Elem().String())
// 		}
// 		res[i] = &out[i]
// 	}

// 	return res, nil
// }

func (me *ArgumentStrategy) ValidateResponseTypes(out []reflect.Type) error {

	if len(out) != len(me.args)+1 {
		return terrors.Errorf("invalid return signature, expected 2, got %d", len(out))
	}

	for i, v := range me.args {
		if !out[i].Implements(reflect.TypeOf(v).Elem()) {
			return terrors.Errorf("invalid return type %q, expected %q", out[i].String(), reflect.TypeOf(v).Elem().String())
		}
	}

	if !out[len(out)-1].Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return terrors.Errorf("invalid return type %q, expected %q", out[len(out)-1].String(), reflect.TypeOf((*error)(nil)).Elem().String())
	}

	return nil
}

func New1ArgumentStrategy[A any]() *ArgumentStrategy {
	return &ArgumentStrategy{args: []any{(*A)(nil)}}
}

func New2ArgumentStrategy[A any, B any]() *ArgumentStrategy {
	return &ArgumentStrategy{args: []any{(*A)(nil), (*B)(nil)}}
}

func New3ArgumentStrategy[A any, B any, C any]() *ArgumentStrategy {
	return &ArgumentStrategy{args: []any{(*A)(nil), (*B)(nil), (*C)(nil)}}
}

func New4ArgumentStrategy[A any, B any, C any, D any]() *ArgumentStrategy {
	return &ArgumentStrategy{args: []any{(*A)(nil), (*B)(nil), (*C)(nil), (*D)(nil)}}
}
