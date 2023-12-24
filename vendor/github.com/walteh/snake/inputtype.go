package snake

import (
	"fmt"
	"time"
)

type StringInput = simpleValueInput[string]
type IntInput = simpleValueInput[int]
type BoolInput = simpleValueInput[bool]
type StringArrayInput = simpleValueInput[[]string]
type IntArrayInput = simpleValueInput[[]int]
type DurationInput = simpleValueInput[time.Duration]
type StringEnumInput = enumInput

type InputType string

func (me InputType) String() string {
	return string(me)
}

var (
	StringInputType      InputType = InputType("string")
	IntInputType         InputType = InputType("int")
	BoolInputType        InputType = InputType("bool")
	StringArrayInputType InputType = InputType("[]string")
	IntArrayInputType    InputType = InputType("[]int")
	DurationInputType    InputType = InputType("time.Duration")
	StringEnumInputType  InputType = InputType("enum")
	UnknownInputType     InputType = InputType("unknown")
)

func AllInputTypes() []InputType {
	return []InputType{
		StringInputType,
		IntInputType,
		BoolInputType,
		StringArrayInputType,
		IntArrayInputType,
		DurationInputType,
		StringEnumInputType,
	}
}

func (me *simpleValueInput[T]) Type() InputType {
	switch Input(me).(type) {
	case *StringInput:
		return StringInputType
	case *IntInput:
		return IntInputType
	case *BoolInput:
		return BoolInputType
	case *StringArrayInput:
		return StringArrayInputType
	case *IntArrayInput:
		return IntArrayInputType
	case *DurationInput:
		return DurationInputType
	default:
		panic(fmt.Errorf("unknown input type %T", me))
	}
}
