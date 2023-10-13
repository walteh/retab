package snake

import (
	"reflect"

	"github.com/go-faster/errors"
)

var (
	ErrInvalidMethodSignature = errors.New("invalid method signatured")
)

func commandResponseValidationStrategy(out []reflect.Type) error {

	if len(out) != 1 {
		return errors.Wrapf(ErrInvalidMethodSignature, "invalid return signature, expected 1, got %d", len(out))
	}

	if !out[0].Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.Wrapf(ErrInvalidMethodSignature, "invalid return type %q, expected %q", out[0].String(), reflect.TypeOf((*error)(nil)).Elem().String())
	}

	return nil
}

func commandResponseHandleStrategy(out []reflect.Value) (*reflect.Value, error) {

	if !out[0].IsNil() {
		return end_of_chain_ptr, out[0].Interface().(error)
	}

	return end_of_chain_ptr, nil
}

func handleArgumentResponse[I any](out []reflect.Value) (*reflect.Value, error) {

	if !out[1].IsNil() {
		return nil, out[1].Interface().(error)
	}

	if out[0].Type() != reflect.TypeOf((*I)(nil)).Elem() {
		return nil, errors.Wrapf(ErrInvalidMethodSignature, "invalid return type %q, expected %q", out[0].String(), reflect.TypeOf((*I)(nil)).Elem().String())
	}

	return &out[0], nil
}

func validateArgumentResponse[I any](out []reflect.Type) error {

	if len(out) != 2 {
		return errors.Wrapf(ErrInvalidMethodSignature, "invalid return signature, expected 2, got %d", len(out))
	}

	if !out[0].Implements(reflect.TypeOf((*I)(nil)).Elem()) {
		return errors.Wrapf(ErrInvalidMethodSignature, "invalid return type %q, expected %q", out[0].String(), reflect.TypeOf((*I)(nil)).Elem().String())
	}

	if !out[1].Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.Wrapf(ErrInvalidMethodSignature, "invalid return type %q, expected %q", out[1].String(), reflect.TypeOf((*error)(nil)).Elem().String())
	}

	return nil
}
