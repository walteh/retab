package snake

import (
	"reflect"
)

func listOfArgs(typ reflect.Type) []reflect.Type {
	var args []reflect.Type

	for i := 0; i < typ.NumIn(); i++ {
		args = append(args, typ.In(i))
	}

	return args
}

func listOfReturns(typ reflect.Type) []reflect.Type {
	var args []reflect.Type

	for i := 0; i < typ.NumOut(); i++ {
		args = append(args, typ.Out(i))
	}

	return args
}

func getRunMethod(inter any) reflect.Value {
	value := reflect.ValueOf(inter)
	method := value.MethodByName("Run")
	if !method.IsValid() {
		if value.CanAddr() {
			method = value.Addr().MethodByName("Run")
		}
	}

	return method
}

func setBindingWithLock[T any](con *Snake, val T) func() {
	con.runlock.Lock()
	defer con.runlock.Unlock()
	ptr := reflect.ValueOf(val)
	typ := reflect.TypeOf((*T)(nil)).Elem()
	con.bindings[typ.String()] = &ptr
	return func() {
		con.runlock.Lock()
		delete(con.bindings, typ.String())
		con.runlock.Unlock()
	}
}
