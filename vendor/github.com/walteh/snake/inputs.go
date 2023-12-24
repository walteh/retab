package snake

import (
	"reflect"
	"strconv"
	"time"
	"unicode"

	"github.com/walteh/terrors"
)

type Input interface {
	Name() string
	Shared() bool
	Ptr() any
	Parent() UntypedResolver
	SetValue(any) error
	Type() InputType
}

type InputWithOptions interface {
	Options() []string
}

func MethodName(m UntypedResolver) string {
	return reflect.ValueOf(m.Ref()).Type().String()
}

func DependancyInputs(str string, m FMap, enum ...Enum) ([]Input, error) {
	deps, err := DependanciesOf(str, m)
	if err != nil {
		return nil, err
	}

	procd := make(map[any]Input, 0)
	nameReserved := make(map[string]string, 0)
	for _, f := range deps {
		if z := m(f); reflect.ValueOf(z).IsNil() {
			return nil, terrors.Errorf("missing resolver for %q", f)
		} else {
			inp, err := InputsFor(z, enum...)
			if err != nil {
				return nil, err
			}
			for _, v := range inp {
				// if they are references to same value, then no need to worry about potential conflicts
				if _, ok := procd[v.Ptr()]; ok {
					continue
				}
				procd[v.Ptr()] = v
				if z, ok := nameReserved[v.Name()]; ok {
					return nil, terrors.Errorf("multiple inputs named %q [%q, %q]", v.Name(), z, MethodName(v.Parent()))
				}
				nameReserved[v.Name()] = MethodName(v.Parent())
			}
		}
	}

	resp := make([]Input, 0)
	for _, v := range procd {
		resp = append(resp, v)
	}

	return resp, nil
}

func InputsFor(m UntypedResolver, enum ...Enum) ([]Input, error) {
	resp := make([]Input, 0)
	for _, f := range StructFields(m) {

		fld := FieldByName(m, f.Name)

		if !f.IsExported() {
			continue
		}

		if f.Type.Kind() == reflect.Ptr {
			f.Type = f.Type.Elem()
			// return nil, terrors.Errorf("field %q in %T is a pointer type", f.Name, m)
		}

		switch f.Type.Kind() {
		case reflect.String:
			if f.Type.Name() != "string" {
				en, err := NewGenericEnumInput(f, m, enum...)
				if err != nil {
					return nil, err
				}
				resp = append(resp, en)
			} else {
				resp = append(resp, NewSimpleValueInput[string](f, m))
			}
		case reflect.Int:
			resp = append(resp, NewSimpleValueInput[int](f, m))
		case reflect.Bool:
			resp = append(resp, NewSimpleValueInput[bool](f, m))
		case reflect.Array, reflect.Slice:
			if f.Type.Elem().Kind() == reflect.String {
				resp = append(resp, NewSimpleValueInput[[]string](f, m))
				continue
			} else if f.Type.Elem().Kind() == reflect.Int {
				resp = append(resp, NewSimpleValueInput[[]int](f, m))
				continue
			}
			return nil, terrors.Errorf("field %q in %v is unexpected reflect.Kind %s", f.Name, m, f.Type.Kind().String())
			// resp = append(resp, NewSimpleValueInput[[]string](f, m))
		case reflect.Int64:
			switch fld.Interface().(type) {
			case time.Duration:
				resp = append(resp, NewSimpleValueInput[time.Duration](f, m))
			default:
				resp = append(resp, NewSimpleValueInput[int64](f, m))
			}
		case reflect.Struct:
			return nil, terrors.Errorf("field %q in %v is unexpected reflect.Kind %s", f.Name, m, f.Type.Kind().String())
		default:
			return nil, terrors.Errorf("field %q in %v is unexpected reflect.Kind %s", f.Name, m, f.Type.Kind().String())
		}

	}
	return resp, nil
}

type genericInput struct {
	field  reflect.StructField
	parent UntypedResolver
}

type simpleValueInput[T any] struct {
	*genericInput
	val *T
}

func (me *simpleValueInput[T]) SetValue(v any) error {
	if x, ok := v.(T); ok {
		*me.val = x
		return nil
	}
	return terrors.Errorf("unable to set value %v to %T", v, me.val)
}

type enumInput struct {
	Enum
}

func (me *enumInput) Name() string {
	return me.Enum.DisplayName()
}

func getEnumOptionsFrom(mytype reflect.Type, enum ...Enum) (Enum, error) {
	rawTypeName := mytype.String()
	var sel Enum
	for _, v := range enum {
		if v.RawTypeName() != rawTypeName {
			continue
		}

		sel = v
	}
	if sel == nil {
		return nil, terrors.Errorf("no options for %q", rawTypeName)
	}

	return sel, nil

}

func NewGenericEnumInput(f reflect.StructField, m UntypedResolver, enum ...Enum) (*enumInput, error) {

	mytype := FieldByName(m, f.Name).Type()

	if mytype.Kind() == reflect.Ptr {
		mytype = mytype.Elem()
	}

	opts, err := getEnumOptionsFrom(mytype, enum...)
	if err != nil {
		return nil, err
	}

	return EnumAsInput(opts, NewGenericInput(f, m)), nil
}

func NewSimpleValueInput[T any](f reflect.StructField, m UntypedResolver) *simpleValueInput[T] {
	v := FieldByName(m, f.Name)

	inp := &simpleValueInput[T]{
		genericInput: NewGenericInput(f, m),
		val:          v.Addr().Interface().(*T),
	}

	return inp
}

func NewGenericInput(f reflect.StructField, m UntypedResolver) *genericInput {
	return &genericInput{
		field:  f,
		parent: m,
	}
}

func (me *simpleValueInput[T]) Value() *T {
	return me.val
}

func (me *genericInput) Name() string {
	// Convert CamelCase (e.g., "NumberOfCats") to kebab-case (e.g., "number-of-cats")
	var result []rune
	for i, r := range me.field.Name {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func (me *genericInput) Shared() bool {
	return MenthodIsShared(me.parent)
}

func (me *genericInput) Parent() UntypedResolver {
	return me.parent
}

func (me *genericInput) Usage() string {
	return me.field.Tag.Get("usage")
}

func (me *simpleValueInput[T]) Ptr() any {
	return me.val
}

func (me *genericInput) Default() string {
	return me.field.Tag.Get("default")
}

func (me *simpleValueInput[T]) Default() T {

	defstr := me.genericInput.Default()

	if defstr == "" && reflect.ValueOf(me.val).IsValid() {
		return *me.val
	}

	switch any(me.val).(type) {
	case *string:
		return any(defstr).(T)
	case *int, *int64:
		if defstr == "" {
			return any(0).(T)
		}
		intt, err := strconv.Atoi(defstr)
		if err != nil {
			panic(err)
		}
		return any(intt).(T)
	case *bool:
		if defstr == "" {
			return any(false).(T)
		}
		boolt, err := strconv.ParseBool(defstr)
		if err != nil {
			panic(err)
		}
		return any(boolt).(T)
	case *time.Duration:
		if defstr == "" {
			return any(time.Second).(T)
		}
		durt, err := time.ParseDuration(defstr)
		if err != nil {
			panic(err)
		}
		return any(durt).(T)
	default:
		panic(terrors.Errorf("unknown type %T", me.val))
	}
}
