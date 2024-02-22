package callgo

import (
	"context"
	"fmt"
	"reflect"
)

var (
	funcs = make(map[string]*Function)

	// precomputed types
	errorType      = reflect.TypeOf((*error)(nil)).Elem()
	stdContextType = reflect.TypeOf((*context.Context)(nil)).Elem()
)

type Function struct {
	fv  reflect.Value // Kind() == reflect.Func
	key string
}

func Func(key string, i any) *Function {
	f := &Function{
		key: key,
		fv:  reflect.ValueOf(i),
	}

	t := f.fv.Type()
	if t.Kind() != reflect.Func {
		panic("callgo: not a function")
	}
	if t.NumIn() != 1 && t.NumIn() != 2 {
		panic("callgo: function must have 1 or 2 arguments")
	}
	if t.In(0) != stdContextType {
		panic("first argument must be context.Context")
	}

	if t.NumOut() != 1 && t.NumOut() != 2 {
		panic("callgo: function must have 1 or 2 return values")
	}
	switch t.NumOut() {
	case 1:
		if t.Out(0) != errorType {
			panic("callgo: single return value must be error")
		}
	case 2:
		if t.Out(1) != errorType {
			panic("callgo: second return value must be error")
		}
	}

	if old := funcs[f.key]; old != nil {
		panic(fmt.Sprintf("callgo: multiple functions registered for task %q", key))
	}
	funcs[f.key] = f

	return f
}
