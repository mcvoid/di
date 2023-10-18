// Package di implements a simple dependency injection library, allowing a program
// to bind dependencies to the different components by type. It is specifically designed
// as a library, not a framework (no inversion of control) to allow maximum flexibility.
// It defines a constructor to make a new injector context and methods to add dependencies
// to the context and to bind them to other objects.
package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

const methodName = "Bind"

var (
	// Returned when the target is nil
	ErrNilInjectee = errors.New("cannot inject into nil value")
	// Returned when the target is not injectable (it is not a function and does not have a bind method)
	ErrNotInjectable = errors.New("is not a function and does not have a 'Bind' method")
	// Returned when it is ambiguous which dependency should be injected (target is an interface which more than one dependency implements)
	ErrAmbiguous = errors.New("more than one dependency implements the interface")
)

// Context is a set of dependencies which can be injected into a bindable object.
type Context struct {
	lock sync.Mutex
	deps map[reflect.Type]reflect.Value
}

// New creates a new Context.
func New() *Context {
	return &Context{
		deps: map[reflect.Type]reflect.Value{},
	}
}

// Add registers a new dependency to the context. If a nil value is passed, that dependency is ignored and no action is taken.
// Dependencies are indexed by type. If two dependencies of the same type are added, the second one overwrites the first.
func (ctx *Context) Add(deps ...interface{}) *Context {
	// Don't change the list while injecting
	// or while adding in another goroutine
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	if ctx.deps == nil {
		ctx.deps = map[reflect.Type]reflect.Value{}
	}

	for _, dep := range deps {

		// nil deps are a no-op
		if dep == nil {
			return ctx
		}

		v := reflect.ValueOf(dep)
		t := v.Type()
		ctx.deps[t] = v
	}

	return ctx
}

// Inject injects the set of dependencies into a bindable object. Can be called on a function or any value with a method called Bind.
// Returns nil if the binding was successful, nil otherwise.
//
// On a function: Calls the function, populating the arguments with values previously added to the Context. The function's return
// value, if any, is discarded.
//
// On an object with a Bind method: Calls the Bind method, populating the arguments with values previously added to the Context. The
// function's return value, if any, is discarded.
//
// Dependencies are bound according to the following rules:
//
//   - If the parameter type is an exact match to a dependency added to the context, that value is used.
//   - If the parameter type is an interface which exactly one dependency implements, that value is used.
//   - If the parameter type is an interface which no dependencies implement, an error is not returned, but rather the argument will
//     be the zero value of the parameter type.
//   - If the parameter type is an interface which more than one dependency implements, an error is returned.
//
// If an error is returned, the function or method is not invoked.
func (ctx *Context) Inject(target interface{}) error {
	if target == nil {
		return ErrNilInjectee
	}
	val := reflect.ValueOf(target)
	t := val.Type()

	if val.Kind() == reflect.Func {
		return injectFunc(ctx, val, t)
	}

	method := val.MethodByName(methodName)
	if method.IsValid() && !method.IsZero() {
		methodType := method.Type()
		return injectFunc(ctx, method, methodType)
	}

	return fmt.Errorf("%w: %v", ErrNotInjectable, target)
}

func injectFunc(ctx *Context, fn reflect.Value, t reflect.Type) error {
	// don't let the list change while we're iterating
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	// iterate the parameters
	// All code paths leading here already validated
	// that the Kind is Func, so no need to worry about panic
	numParams := t.NumIn()
	in := make([]reflect.Value, numParams)
	for i := 0; i < numParams; i++ {
		argType := t.In(i)
		if val, ok := ctx.deps[argType]; ok {
			in[i] = val
			continue
		}

		// can't find a one-to-one type match
		// do a search and find everything that
		// implements the requested type
		candidateVals := []reflect.Value{}
		candidateTypes := []reflect.Type{}
		for t, val := range ctx.deps {
			if t.Implements(argType) {
				candidateVals = append(candidateVals, val)
				candidateTypes = append(candidateTypes, t)
			}
		}

		// no matches means we pass zero
		if len(candidateVals) == 0 {
			in[i] = reflect.Zero(argType)
			continue
		}

		// too many matches
		if len(candidateVals) > 1 {
			return fmt.Errorf("%w, bound types with possible match: %v", ErrAmbiguous, candidateTypes)
		}

		// exactly one match - perfect
		in[i] = candidateVals[0]
	}

	fn.Call(in)
	return nil
}
