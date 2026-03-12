# Control Flow

Avenir provides several control flow constructs: conditionals, loops, and exception handling.

## Conditionals

### If Statements

If statements execute code conditionally:

```avenir
if (condition) {
    // code
}
```

The condition must be of type `bool`.

### If-Else Statements

Else clauses provide alternative execution paths:

```avenir
if (condition) {
    // code
} else {
    // alternative code
}
```

### Else-If Chaining

Multiple conditions can be chained using else-if:

```avenir
if (condition1) {
    // code
} else if (condition2) {
    // code
} else {
    // default code
}
```

### Multiple Conditions

If statements support multiple conditions using semicolons:

```avenir
if (x > 0; x < 10) {
    // equivalent to: if (x > 0 && x < 10)
}
```

## Loops

### While Loops

While loops execute code while a condition is true:

```avenir
while (condition) {
    // code
}
```

The condition must be of type `bool`.

### For Loops

For loops provide C-style iteration:

```avenir
for (init; condition; post) {
    // code
}
```

All parts (init, condition, post) are optional:

```avenir
for (; condition; ) { }    // while loop equivalent
for (;;) { }               // infinite loop
```

The init can be a variable declaration or an assignment:

```avenir
for (var i | int = 0; i < 10; i = i + 1) {
    // code
}
```

### For-Each Loops

For-each loops iterate over lists:

```avenir
for (item in list) {
    // code
}
```

The loop variable `item` is scoped to the loop body. The list expression must be of type `list<T>`.

Example:

```avenir
var numbers | list<int> = [1, 2, 3];
for (n in numbers) {
    print(n);
}
```

### Break Statements

Break statements exit the innermost loop:

```avenir
while (true) {
    if (condition) {
        break;
    }
}
```

Break can only be used inside loops.

### Continue Statements

Continue statements skip the rest of the current loop iteration and proceed with
the next one:

```avenir
for (var i | int = 0; i < 5; i = i + 1) {
    if (i % 2 == 0) {
        continue;
    }
    print(i);
}
```

Continue can only be used inside loops.

## Switch Statements

Switch statements perform value-based branching with explicit `case` clauses and
an optional `default` clause:

```avenir
switch code {
    case 200:
        print("ok");
    case 404:
        print("not found");
    default:
        print("other");
}
```

Cases are matched by equality (`==`). Fallthrough is not supported.

## Exception Handling

### Try-Catch Statements

Try-catch statements handle exceptions:

```avenir
try {
    // code that may throw
} catch (e | error) {
    // handle exception
}
```

The catch variable can be of type `error` or a struct type. The catch block is
executed if an exception is thrown in the try block, including runtime and
builtin errors.

### Typed Catch Clauses

Try statements support multiple typed catch clauses for matching specific error
types. Clauses are evaluated in order; the first matching clause handles the
error:

```avenir
struct FileNotFound {
    path | string;
}

struct PermissionDenied {
    file | string;
}

fun riskyOp() | void ! FileNotFound, PermissionDenied {
    throw FileNotFound{path = "/tmp/missing.txt"};
}

fun main() | void {
    try {
        riskyOp();
    } catch (e | FileNotFound) {
        print("not found: " + e.path);
    } catch (e | PermissionDenied) {
        print("denied: " + e.file);
    } catch (e | error) {
        print("other error");
    }
}
```

A `catch (e | error)` clause acts as a fallback that catches any error not
matched by preceding clauses.

### Throw Statements

Throw statements raise exceptions:

```avenir
throw error("something went wrong");
```

The expression can be of type `error` or a declared struct type.

### Throws Declarations

Functions can declare which error types they may throw using the `!` syntax
after the return type:

```avenir
fun readFile(path | string) | string ! FileNotFound {
    throw FileNotFound{path = path};
}
```

Multiple thrown types are separated by commas:

```avenir
fun process() | void ! FileNotFound, PermissionDenied {
    // ...
}
```

The type checker validates that `throw` expressions inside the function body
match the declared throws types.

## Deferred Calls

`defer` registers a call expression to run when the current function returns.
Deferred calls are executed in LIFO order.

```avenir
fun cleanupDemo() | void {
    defer log("first");
    defer log("second");
    log("body");
}
```

The output order is: `body`, `second`, `first`.

Current limitation: only call expressions are supported in `defer`.

### Exception Propagation

If an exception is not caught, it propagates up the call stack. If it reaches the top level, the program terminates with an error.

## Block Statements

Blocks create new scopes:

```avenir
{
    var x | int = 10;
    // x is available here
}
// x is not available here
```

Blocks are used in conditionals, loops, and functions.

## Expression Statements

Expressions can be used as statements:

```avenir
print("Hello");
x + y;  // Expression statement (result is discarded)
```
