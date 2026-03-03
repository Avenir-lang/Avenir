# Frequently Asked Questions

## General

### What is Avenir?

Avenir is a statically-typed programming language designed for building reliable software. It features a clean syntax, strong type safety, and a virtual machine runtime.

### What platforms does Avenir support?

Avenir is implemented in Go and runs on platforms supported by Go (Linux, macOS, Windows, etc.).

### Is Avenir production-ready?

Avenir is under active development. Some features may not be fully implemented or may change.

## Language Features

### Does Avenir support generics?

Not currently. Generic types are not yet supported.

### Does Avenir support inheritance?

No. Avenir uses composition and methods, not inheritance.

### Does Avenir support interfaces?

Yes. Avenir supports structural typing with interfaces. A type implements an interface implicitly if it has all required methods with matching signatures. See [Types](types.md) for details.

### Does Avenir support operator overloading?

No. Operators have fixed semantics.

### Does Avenir support function overloading?

No. Each function name must be unique within a scope.

### Does Avenir support optional chaining?

Not currently. Optional types exist, but optional chaining is not implemented.

### Does Avenir support pattern matching?

Not currently. Pattern matching is not implemented.

## Type System

### Are types inferred?

No. Avenir requires explicit type annotations. Type inference is not supported.

### Can I use `any` everywhere?

While `any` allows any type, it's generally better to use specific types for type safety.

### Are structs value types or reference types?

Structs are value types. Assignment creates a copy. However, mutable fields are mutated in-place.

### How do I handle nullable values?

Use optional types (`T?`) with `some()` and `none()`:

```avenir
var x | int? = some(42);
var y | int? = none;
```

## Modules and Imports

### How do I create a module?

Create a file with a package declaration:

```avenir
pckg mymodule;

// Your code here
```

### How do I import a module?

Use the `import` statement:

```avenir
import mymodule;
```

### What is the file-to-struct mapping rule?

If a file contains structs, **at least one struct must have the same name as the file** (without the `.av` extension). For example:

- `Point.av` must contain `struct Point` (if it contains any structs)
- Files without structs can still be imported (for function-only modules)

### How do folder-based imports work?

If a folder `A` contains a file `A.av`, you can import it using just the folder name:

```
Folder: geometry/
File: geometry/geometry.av
Import: import geometry;
```

The compiler automatically resolves `import geometry` to `geometry/geometry.av`.

### Can I have circular imports?

No. Import cycles are detected and reported as errors.

### Where are modules located?

Modules are resolved based on file paths:

- **Standard modules** (`std.*`): Looked up in the `std/` directory
- **Application modules**: Looked up relative to the project root

The compiler tries folder-based imports first (`module/A/A.av`), then falls back to flat files (`module/A.av`).

## Functions and Methods

### What's the difference between a function and a method?

A function is a standalone function. A method is a function associated with a type (instance or static).

### Can I have default parameters?

Yes. Function parameters can have default values:

```avenir
fun greet(name | string, greeting | string = "Hello") | void {
    // ...
}
```

### Can I use named arguments?

Yes. Function calls support named arguments:

```avenir
create_point(x = 10, y = 20);
```

### Can I return multiple values?

Not directly. Use a struct or list to return multiple values:

```avenir
struct Result {
    value | int
    error | error?
}

fun compute() | Result {
    return Result{value = 42, error = none};
}
```

## Control Flow

### Can I use `break` in if statements?

No. `break` can only be used in loops.

### Can I use `continue`?

Not currently. `continue` is not implemented.

### Can I use `switch` statements?

Not currently. Use if-else chains instead.

### Can I use `defer`?

Not currently. `defer` is not implemented.

## Error Handling

### How do I handle errors?

Use try-catch blocks:

```avenir
try {
    // code that may throw
} catch (e | error) {
    // handle error
}
```

### Can I return errors instead of throwing?

Not currently. Errors must be thrown as exceptions.

### Can I create custom error types?

Not currently. All errors must be of type `error`.

## Performance

### Is Avenir fast?

The VM is optimized for execution, but compiled native code would be faster (not yet implemented).

### How does Avenir compare to other languages?

Avenir is designed for reliability and clarity, with performance considerations but not as the primary goal.

## Development

### How do I debug Avenir programs?

Use `print` statements for debugging. A debugger is not currently available.

### How do I test Avenir programs?

Write test programs and run them. A testing framework is not currently available.

### How do I profile Avenir programs?

Profiling tools are not currently available.

## Contributing

### How can I contribute?

See the project repository for contribution guidelines.

### Where can I report bugs?

Report bugs in the project issue tracker.

### Where can I ask questions?

See the project repository for discussion forums.
