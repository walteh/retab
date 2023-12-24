package snake

import (
	"reflect"
	"slices"
	"strings"

	"github.com/walteh/terrors"
)

func ResolveEnum[T ~string](name string, options []T, resolver EnumResolverFunc) (T, error) {
	if resolver == nil {
		return "", terrors.Errorf("no enum resolver for %q", reflect.TypeOf((*T)(nil)).Elem().String())
	}

	if len(options) == 0 {
		return "", terrors.Errorf("no options for %q", reflect.TypeOf((*T)(nil)).Elem().String())
	}

	check := T(name)

	strs := make([]string, len(options))
	for i, v := range options {
		strs[i] = string(v)
	}

	if name == "select" {

		resp, err := resolver(reflect.TypeOf((*T)(nil)).Elem().String(), strs)
		if err != nil {
			return "", err
		}

		check = T(resp)

	}

	if !slices.Contains(options, check) {
		return "", terrors.Errorf("invalid value %q, expected one of [\"%s\"]", check, strings.Join(strs, "\", \""))
	}

	return T(check), nil
}

// REFRESHABLE RESOLVER
var (
	_ UntypedResolver = (*rawEnum[string])(nil)
)

type EnumResolverFunc func(typeName string, options []string) (string, error)

type Enum interface {
	UntypedResolver
	Input
	SetCurrent(string) error
	CurrentPtr() *string
	RawTypeName() string
	Options() []string
	Ptr() any
	Usage() string
	DisplayName() string
	ApplyResolver(EnumResolverFunc) error
}

type rawEnum[T ~string] struct {
	rawTypeName  string
	options      []T
	enumResolver EnumResolverFunc
	name         string
	description  string
	// Val needs to be exported value so it is picked up in inputs.go reflection logic
	Val *T
}

// Name implements Enum.
func (me *rawEnum[T]) Name() string {
	return me.name
}

// Parent implements Enum.
func (m *rawEnum[T]) Parent() UntypedResolver {
	return m
}

// Shared implements Enum.
func (*rawEnum[T]) Shared() bool {
	return true
}

func (me *rawEnum[T]) Usage() string {
	return me.description
}

// because we know the run method is correct
func (me *rawEnum[T]) isRund() {}

// Ref implements ValidatedRunMethod.
func (me *rawEnum[T]) Ref() Method {
	return me
}

// RunFunc implements ValidatedRunMethod.
func (me *rawEnum[T]) RunFunc() reflect.Value {
	return reflect.ValueOf(me.Run)
}

func NewEnumOptionWithResolver[T ~string](name string, description string, input ...T) Enum {
	sel := new(T)

	return &rawEnum[T]{
		rawTypeName: reflect.TypeOf((*T)(nil)).Elem().String(),
		options:     input,
		name:        name,
		description: description,
		Val:         sel,
	}
}

func (me *rawEnum[T]) SetValue(v any) error {
	if x, ok := v.(string); ok {
		return me.SetCurrent(x)
	}
	return terrors.Errorf("unable to set value %v to %T", v, me.Val)
}

func (me *rawEnum[T]) ApplyResolver(resolver EnumResolverFunc) error {
	me.enumResolver = resolver
	*me.Val = T("select")
	return nil
}

func (me *rawEnum[T]) DisplayName() string {
	return me.name
}

func (me *rawEnum[T]) RawTypeName() string {
	return me.rawTypeName
}

func (me *rawEnum[T]) Description() string {
	return me.description
}

func (me *rawEnum[T]) OptionsWithSelect() []string {
	opts := me.Options()
	if me.enumResolver != nil {
		opts = append(opts, "select")
	}
	return opts
}

func (me *rawEnum[T]) Options() []string {
	opts := make([]string, len(me.options))
	for i, v := range me.options {
		opts[i] = string(v)
	}
	return opts
}

func (e *rawEnum[I]) SetCurrent(vt string) error {
	if slices.Contains(e.OptionsWithSelect(), string(vt)) {
		*e.Val = I(vt)
		return nil
	}
	return terrors.Errorf("invalid value %q, expected one of [\"%s\"]", vt, strings.Join(e.OptionsWithSelect(), "\", \""))
}

func (e *rawEnum[I]) CurrentPtr() *string {
	return (*string)(reflect.ValueOf(e.Val).UnsafePointer())
}

func (me *rawEnum[T]) Run() (T, error) {
	if me.Val == nil || reflect.ValueOf(me.Val).IsNil() || *me.Val == "select" {
		if me.enumResolver == nil {
			return "", terrors.Errorf("no enum resolver for %q", me.rawTypeName)
		}

		resolve, err := me.enumResolver(me.rawTypeName, me.Options())
		if err != nil {
			return "", err
		}

		if err := me.SetCurrent(resolve); err != nil {
			return "", err
		}
	}
	return *me.Val, nil
}

func EnumAsInput(me Enum, m *genericInput) *enumInput {
	return &enumInput{
		Enum: me,
	}
}

func (me *rawEnum[I]) Ptr() any {
	return me.CurrentPtr()
}

func (me *rawEnum[I]) IsResolver() {}

func (me *rawEnum[I]) Type() InputType {
	return StringEnumInputType
}
