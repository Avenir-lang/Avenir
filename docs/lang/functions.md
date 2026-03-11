# Functions

Functions in Avenir are first-class values. Functions can be named or anonymous (function literals).

## Function Declaration

Functions are declared using the `fun` keyword:

```avenir
fun name(param1 | type1, param2 | type2) | returnType {
    // body
}
```

The return type is required. Use `void` if the function doesn't return a value.

### Async Functions

Functions can be declared as async with the `async` keyword:

```avenir
async fun fetchNumber() | int {
    return 10;
}
```

Calling an async function automatically spawns a concurrent task and returns a `Future<T>`. The caller can store the future or immediately `await` it.

### Await Expressions

Use `await` inside async functions to get the resolved value:

```avenir
async fun main() | int {
    var a | int = await fetchNumber();
    return a + 1;
}
```

`await` suspends the current task until the future resolves. Other tasks may run while suspended.

### Concurrent Spawn

Calling multiple async functions before awaiting enables true concurrency:

```avenir
async fun compute(x | int) | int {
    await asyncSleep(50000000);
    return x * 2;
}

async fun main() | int {
    var a | Future<int> = compute(10);
    var b | Future<int> = compute(20);
    var ra | int = await a;
    var rb | int = await b;
    return ra + rb;
}
```

Both `compute` calls run concurrently. The total time is ~50ms, not ~100ms.

### Generic Functions

Functions can declare type parameters after the function name:

```avenir
fun identity<T>(x | T) | T {
    return x;
}

fun pickFirst<T, U>(a | T, b | U) | T {
    return a;
}
```

At call sites, generic type arguments are explicit:

```avenir
var n | int = identity<int>(42);
var s | string = identity<string>("hello");
```

Type argument inference is not supported yet.

Example:

```avenir
fun add(x | int, y | int) | int {
    return x + y;
}

fun print_hello() | void {
    print("Hello");
}
```

## Decorators

Decorators allow wrapping functions with additional behavior using the `@<expr>` syntax:

```avenir
@log
fun add(a | int, b | int) | int {
    return a + b;
}
```

A decorator is any expression that evaluates to a function taking a function as its argument and returning a function of the same type. The above is equivalent to `add = log(add)`.

Decorators are applied **once at module initialization**, not on every function call. This means decorated functions have **zero runtime overhead** — the original function slot is permanently replaced with the decorated version before `main` executes.

### Decorator Requirements

The decorator function must:
- Accept exactly one parameter: the function being decorated
- Return a function with the same signature as the decorated function

```avenir
fun doubler(f | fun(int, int) | int) | fun(int, int) | int {
    return fun(a | int, b | int) | int {
        return f(a, b) * 2;
    };
}

@doubler
fun add(a | int, b | int) | int {
    return a + b;
}

// add(3, 4) now returns 14 instead of 7
```

### Parameterized Decorators

Decorators can take arguments using parentheses. A parameterized decorator is a function that returns a decorator:

```avenir
@cache(60)
fun compute(x | int) | int {
    return x * x;
}
```

Here `cache(60)` is called first, and the returned function is used as the decorator.

### Decorators and Methods

Decorators work with both instance methods and static methods.

For instance methods, the method type includes the receiver as the first parameter.
So the decorator must accept and return that full function type:

```avenir
struct Point {
    x | int
    y | int
}

fun method_doubler(f | fun(Point, int, int) | int) | fun(Point, int, int) | int {
    return fun(self | Point, dx | int, dy | int) | int {
        return f(self, dx, dy) * 2;
    };
}

@method_doubler
fun (self | Point).move_score(dx | int, dy | int) | int {
    return self.x + self.y + dx + dy;
}

// Static method with decorator
@validate
fun Point.origin() | Point {
    return Point{x = 0, y = 0};
}
```

### Static Methods as Decorators

Static methods can be used as decorators since they are statically accessible:

```avenir
struct Logger {
}

fun Logger.log(f | fun() | void) | fun() | void {
    return fun() | void {
        print("Starting function");
        f();
        print("Function completed");
    };
}

fun Logger.validate(f | fun(int) | int) | fun(int) | int {
    return fun(x | int) | int {
        if x < 0 {
            print("Warning: negative input");
        }
        return f(x);
    };
}

// Using static methods as decorators
@Logger.log
fun test() | void {
    print("Hello, World!");
}

@Logger.validate
fun process(x | int) | int {
    return x * 2;
}
```

**Note:** Since decorators are arbitrary expressions, any expression that evaluates to a decorator function can be used, including static method access like `Logger.log`. Instance methods cannot be decorators because they require a specific instance.

**Note:** Decorated methods are declared at top level with receiver syntax (`fun (self | T).name(...) ...`).
They are not declared inside the `struct { ... }` field block.

### Multiple Decorators

Multiple decorators can be stacked. They are applied bottom-up (innermost first):

```avenir
@dec1
@dec2
fun f(x | int) | int {
    return x;
}
// Equivalent to: f = dec1(dec2(f))
```

### Generic Decorators

Decorators can be generic functions using variadic type parameters for universal applicability:

```avenir
fun wrap<R, ...Args>(f | fun(Args...) | R) | fun(Args...) | R {
    return f;
}
```

When applied to a function, the type arguments are inferred automatically from the decorated function's signature.

Usage:

```avenir
@wrap
fun add(a | int, b | int) | int {
    return a + b;
}

@wrap
fun greet(name | string, times | int) | string {
    return "hello";
}
```

In these examples, `wrap` is instantiated with different `(R, Args...)` based on each decorated function.

## Variadic Generics

Variadic type parameters allow functions to accept a variable number of type arguments.

### Variadic Type Parameters

Declare a variadic type parameter with `...` before the name:

```avenir
fun wrap<R, ...Args>(f | fun(Args...) | R) | fun(Args...) | R {
    return f;
}
```

- `...Args` declares `Args` as a variadic type parameter (a type pack)
- `Args...` expands the type pack in type position

### Type Pack Expansion

Use `Name...` in type positions to expand a type pack:

```avenir
fun(Args...) | R    // function taking the expanded types and returning R
```

When `wrap` is applied to a `fun(int, string) | bool`, the type checker infers:
- `R = bool`
- `Args = (int, string)`

And `fun(Args...) | R` expands to `fun(int, string) | bool`.

## Public Functions

Functions can be marked as public using the `pub` keyword:

```avenir
pub fun exported_function() | void {
    // This function is accessible from other modules
}
```

Private functions (without `pub`) are only accessible within the same module.

## Parameters

Function parameters must have explicit type annotations:

```avenir
fun greet(name | string, age | int) | void {
    print(name);
    print(age);
}
```

### Default Parameters

Parameters can have default values:

```avenir
fun greet(name | string, greeting | string = "Hello") | void {
    print(greeting);
    print(name);
}
```

Default values must be compile-time constants.

When calling a function with default parameters, you can omit arguments:

```avenir
greet("Alice");              // greeting defaults to "Hello"
greet("Bob", "Hi");          // greeting is "Hi"
```

### Named Arguments

Function calls support named arguments:

```avenir
fun create_point(x | int, y | int) | Point {
    return Point{x = x, y = y};
}

create_point(x = 10, y = 20);
create_point(y = 20, x = 10);  // Order doesn't matter
```

Named arguments can be mixed with positional arguments, but positional arguments must come before named arguments:

```avenir
create_point(10, y = 20);      // OK
create_point(x = 10, 20);      // Error: positional after named
```

## Return Values

Functions return values using the `return` statement:

```avenir
fun add(x | int, y | int) | int {
    return x + y;
}
```

Functions with `void` return type can use `return` without a value:

```avenir
fun early_exit() | void {
    if (condition) {
        return;
    }
    // ...
}
```

## Function Literals

Function literals (anonymous functions) are supported:

```avenir
var add | fun (int, int) | int = fun (x | int, y | int) | int {
    return x + y;
};
```

Function literals can be passed as arguments and stored in variables.

## Closures

Function literals can capture variables from their enclosing scope (closures):

```avenir
fun make_counter() | fun () | int {
    var count | int = 0;
    return fun () | int {
        count = count + 1;
        return count;
    };
}
```

Closures capture variables by reference, allowing them to modify captured variables.

## Function Types

Function types are specified using the `fun` keyword:

```avenir
var fn | fun (int, int) | int = add;
```

Function types use structural typing: two function types are equal if they have the same parameter types and return type.

## Function Calls

Functions are called using parentheses:

```avenir
var result | int = add(10, 20);
```

Generic calls include type arguments:

```avenir
var value | int = identity<int>(10);
```

Method calls use dot notation:

```avenir
var length | int = str.length();
```

## Main Function

Every program must have a `main` function that serves as the entry point:

```avenir
fun main() | void {
    // Program entry point
}
```

The `main` function is called automatically when the program runs.

`main` can also be async:

```avenir
async fun main() | int {
    var a | int = await fetchNumber();
    return a;
}
```
