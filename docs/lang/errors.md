# Error Handling

Avenir uses exceptions for error handling. Exceptions are represented by the `error` type and are thrown for:

- Explicit `throw` statements
- Runtime failures (division by zero, invalid access, etc.)
- Builtin failures (e.g., invalid string conversions)

## Error Type

The `error` type represents exceptional conditions:

```avenir
var e | error = error("something went wrong");
```

Errors are created using the `error()` built-in function. Runtime and builtin errors are created automatically by the VM and surfaced as `error` values.

## Throwing Exceptions

Exceptions are thrown using the `throw` statement:

```avenir
throw error("something went wrong");
```

The expression can be of type `error` or a declared struct type.

## Typed Errors

Functions can declare which error types they may throw using the `!` syntax
after the return type:

```avenir
struct FileNotFound {
    path | string
}

fun readFile(path | string) | string ! FileNotFound {
    throw FileNotFound{path = path};
}
```

Multiple thrown types are separated by commas:

```avenir
struct PermissionDenied {
    operation | string
}

fun process() | void ! FileNotFound, PermissionDenied {
    // ...
}
```

The type checker validates that `throw` expressions inside the function body
match the declared throws types.

## Catching Exceptions

Exceptions are caught using try-catch blocks:

```avenir
try {
    // code that may throw
    throw error("error occurred");
} catch (e | error) {
    // handle exception
    print(errorMessage(e));
}
```

### Multiple Typed Catch Clauses

You can catch different error types with multiple catch clauses:

```avenir
struct FileNotFound {
    path | string
}

struct PermissionDenied {
    operation | string
}

fun processFile(path | string) | void ! FileNotFound, PermissionDenied {
    // ... code that may throw different error types
}

try {
    processFile("/data.txt");
} catch (e | FileNotFound) {
    print("File not found: ${e.path}");
} catch (e | PermissionDenied) {
    print("Permission denied for: ${e.operation}");
} catch (e | error) {
    print("Generic error: ${errorMessage(e)}");
}
```

The catch clauses are tried in order. The first matching type is executed, and subsequent clauses are skipped. The final `catch (e | error)` clause acts as a fallback for any remaining error types.

Runtime and builtin errors are also catchable:

```avenir
try {
    var n | int = toInt("abc"); // builtin error
    var x | int = 10 / 0;       // runtime error
    print(n + x);
} catch (e | error) {
    print(errorMessage(e));
}
```

The catch variable can be of type `error` or a struct type. The catch block is executed if an exception is thrown in the try block, including runtime and builtin errors.

## Error Messages

Error messages are extracted using the `errorMessage()` built-in function:

```avenir
var msg | string = errorMessage(e);
```

## Exception Propagation

If an exception is not caught, it propagates up the call stack:

```avenir
fun f() | void {
    throw error("error in f");
}

fun g() | void {
    f();  // Exception propagates from f to g
}

fun main() | void {
    try {
        g();  // Exception propagates from g to main
    } catch (e | error) {
        // Exception is caught here
    }
}
```

If an exception reaches the top level (e.g., in `main`) and is not caught, the program terminates with an error.

## Unhandled Exceptions

Unhandled exceptions cause program termination:

```avenir
fun main() | void {
    throw error("unhandled");
    // Program terminates here
}
```

## Error Handling Patterns

### Propagating Errors

Functions can propagate errors by not catching them:

```avenir
fun may_fail() | void {
    if (condition) {
        throw error("failure");
    }
}
```

### Handling Errors

Functions can handle errors using try-catch:

```avenir
fun safe_operation() | void {
    try {
        may_fail();
    } catch (e | error) {
        // Handle error
        print(errorMessage(e));
    }
}
```

### Returning Errors

Currently, Avenir uses exceptions for error handling. Functions can return `error` values, but exceptions are thrown explicitly with `throw` or implicitly by runtime/builtin failures.

## Error Type Restrictions

The `error` type is a built-in type with a message and extensible metadata. User-defined error types are supported as struct types that can be thrown and caught. All exceptions must be of type `error` or a struct type.

When using typed errors:
- Struct types can be thrown and caught specifically
- The `error` type catches all exceptions including struct types
- Type checking ensures only declared error types are thrown
- Multiple catch clauses allow fine-grained error handling

## Best Practices

- **Handle errors appropriately**: Don't ignore exceptions
- **Provide clear error messages**: Use descriptive messages in `error()` calls
- **Use try-catch selectively**: Only catch exceptions where you can handle them
- **Document error conditions**: Document which functions may throw
