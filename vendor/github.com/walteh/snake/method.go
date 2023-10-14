package snake

import (
	"reflect"

	"github.com/spf13/pflag"
)

type method struct {
	name   string
	method reflect.Value

	flags              func(*pflag.FlagSet)
	validationStrategy func([]reflect.Type) error
	responseStrategy   func([]reflect.Value) (*reflect.Value, error)
	cmd                Cobrad
}

type Method interface {
	Flags(*pflag.FlagSet)
	Run() reflect.Value
	RunArgs() []reflect.Type
	ValidateResponse() error
	HandleResponse([]reflect.Value) (*reflect.Value, error)
	Name() string
	Command() Cobrad
}

var _ Method = (*method)(nil)

func (me *method) Flags(flags *pflag.FlagSet) {
	me.flags(flags)
}

func (me *method) Run() reflect.Value {
	return me.method
}

func (me *method) RunArgs() []reflect.Type {
	return listOfArgs(me.method.Type())
}

func (me *method) ValidateResponse() error {
	return me.validationStrategy(listOfReturns(me.method.Type()))
}

func (me *method) HandleResponse(out []reflect.Value) (*reflect.Value, error) {
	return me.responseStrategy(out)
}

func (me *method) Name() string {
	return me.name
}

func (me *method) Command() Cobrad {
	return me.cmd
}
