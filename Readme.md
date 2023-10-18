# DI

An extremely simple dependency injection library for Go.

## Install

`$ go get github.com/mcvoid/di`

## Usage

### Hello, world!
The simplest usage is like such:

```
di.New().Add(os.Stdout).Inject(func(w io.Writer) {
  fmt.Fprintf(w, "Hello, world!\n")
})
```

What is it doing?

### Step 1: Create a Context

A context is a set of dependencies which will get injected at once. You might
want to create one context for deployment and one for unit testing, for example.
The deployment context will hold the real versions of your dependencies, while the
test context might hold your mocks.

You can create a context with the new method.

```
ctx := di.New()
```

Or you can just declare it directly - the zero value is a valid context.

```
var ctx di.Context
```

### Step 2: Add Dependencies

Once a context is created, you'll want to add something to the set. Dependencies
are identified by type.

```
ctx.Add(mydep)
```

It's a fluent interface (like a Builder pattern), so you can chain as many
`Add`s together as you like.

```
ctx.Add(mydep1).Add(mydep2).Add(mydep3)
```

Or add several things at once.

```
ctx.Add(mydep1, mydep2, mydep3)
```

Note that since they are identified by type, adding several items of the same type
has the effect of overwriting older items.

### Step 3: Inject

Then it's time to inject the dependencies into an object. There's two ways of doing so:
via a function, and via a method.

#### Function Injection

Injecting into a function will invoke that function, but DI will fill in the function's
parameters according to what's inside the context. You can see that in the Hello World
example above: a `Writer` as added to the context (in this case `os.Stdout`), and the
function asked for a `Writer` in one of its parameters, and DI gave it the one in the
context.

You can ask for as many dependencies as you want, each one being its own parameter.
You don't have to specify an exact type, either - if your parameter is an interface
and something in the context implements that interface, DI will find it and inject
it for you. Like in the Hello, World example: `os.Stdout` is of type `*os.File`, and
the function asked for an `io.Writer`.

Example:

```
func doTheThing(a TypeA, b TypeB, c TypeC) {
  // do things with the dependencies
}

// later, in some function
ctx.Inject(doTheThing)
```

#### Method Injection

Maybe you just need an object to be populated. In that case, DI can inject into any
object that has a `Bind` method. Much like with injecting into a function, DI will
invoke the method and fill in the parameters of the `Bind` method. With that you can
use the method to populate your object.

Example:

```
type myThing struct {
  propA TypeA
  propB TypeB
  propC TypeC
}

func (t *myThing) Bind(a TypeA, b TypeB, c TypeC) {
  t.propA = a
  t.propB = b
  t.propC = c
}

// later, in some function
t := myThing{}
ctx.Inject(&t)
```

### Restrictions

* Any return value of an injected function or method will be dropped.
* Adding nil values to a context is a no-op.
* If no type matches, the parameter will be its zero value.
* If a function or method asks for an interface that is implemented by
more than one dependency in the context, `Inject` will return an error.

### License

DI is released under the standard 3-clause BSD license.
