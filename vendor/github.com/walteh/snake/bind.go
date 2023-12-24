package snake

import (
	"reflect"
	"sync"
)

type Binder struct {
	bindings map[string]*reflect.Value
	runlock  sync.Mutex
}

func (me *Binder) Bound(name string) *reflect.Value {
	me.runlock.Lock()
	defer me.runlock.Unlock()
	return me.bindings[name]
}

func (me *Binder) Bind(name string, val *reflect.Value) {
	me.runlock.Lock()
	defer me.runlock.Unlock()
	me.bindings[name] = val
}

func NewBinder() *Binder {
	return &Binder{
		bindings: make(map[string]*reflect.Value),
	}
}

func SetBinding[T any](con *Binder, val T) func() {
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

func SetBindingIfNil[T any](con *Binder, val T) func() {
	con.runlock.Lock()
	defer con.runlock.Unlock()
	ptr := reflect.ValueOf(val)
	typ := reflect.TypeOf((*T)(nil)).Elem()
	if _, ok := con.bindings[typ.String()]; !ok {
		// check if it is the snake.NewNoopMethod
		con.bindings[typ.String()] = &ptr
		return func() {
		}
	}
	return func() {
		con.runlock.Lock()
		delete(con.bindings, typ.String())
		con.runlock.Unlock()
	}
}

func SetBindingWithLock[T any](con *Binder, val T) func() {
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
