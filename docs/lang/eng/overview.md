# Avenir Language Overview

Avenir is a statically-typed programming language designed for building reliable software. The language combines a clean, readable syntax with strong type safety and a comprehensive runtime system.

## Key Features

### Type System

- **Static typing**: All types are checked at compile time
- **Explicit typing**: Variable types must be explicitly declared (no inference)
- **Structural and nominal types**: Lists use structural typing, structs use nominal typing
- **Union types**: Support for type unions (`<int|string>`)
- **Optional types**: Optional values (`T?`) for nullable types
- **Function types**: First-class function types

### Core Types

- **Primitives**: `int`, `float`, `string`, `bool`, `bytes`
- **Collections**: `list<T>` and `dict<T>`
- **Structs**: User-defined types with named fields
- **Functions**: First-class functions and closures
- **Error**: Built-in error type for exception handling

### Struct System

- **Visibility control**: `pub` keyword for public structs and fields
- **Mutability control**: `mut struct` for mutable structs, field-level `mut` override
- **Default values**: Struct fields can have compile-time default values
- **Methods**: Both instance and static methods
- **Field assignment**: In-place mutation for mutable fields

### Functions

- **Named functions**: Top-level function declarations
- **Function literals**: Anonymous functions and closures
- **Default parameters**: Function parameters can have default values
- **Named arguments**: Function calls support named arguments
- **Methods**: Instance methods with receivers, static methods on types

### Control Flow

- **Conditionals**: `if`/`else` statements with optional chaining
- **Loops**: `while`, C-style `for`, and `for...in` loops
- **Exceptions**: `try`/`catch` for error handling
- **Break**: Early loop termination

### Module System

- **Package declarations**: Every file starts with `pckg` declaration
- **Imports**: Module-based imports with optional aliases
- **File-to-struct mapping**: Files with structs must have matching struct names
- **Folder-based imports**: Support for `A/A.av` → `import module.A` syntax
- **Visibility**: Module-level visibility control with `pub`
- **Module resolution**: Dual resolution (folder-based and flat files)

### Runtime

- **Virtual machine**: Stack-based bytecode VM
- **Built-in functions**: Core functions for I/O, collections, strings, bytes
- **Built-in methods**: Methods on built-in types (strings, lists, bytes)
- **Closures**: Full support for closures and lexical scoping

## Design Philosophy

Avenir is designed for:

- **Type safety**: Catch errors at compile time, not runtime
- **Explicitness**: Clear, readable code with explicit types
- **Practicality**: Features that enable real-world software development
- **Performance**: Compiled to bytecode with an efficient VM

## Example

```avenir
pckg main;

struct Point {
    x | int
    y | int
}

fun (p | Point).sum() | int {
    return p.x + p.y;
}

fun main() | void {
    var p | Point = Point{x = 3, y = 4};
    var s | int = p.sum();
    print(s);
}
```

## Documentation Structure

This documentation covers:

- **Getting Started**: Installation and your first program
- **Syntax**: Language syntax and grammar
- **Types**: Type system and type checking
- **Variables**: Variable declarations and assignments
- **Functions**: Functions, parameters, and closures
- **Control Flow**: Conditionals, loops, and error handling
- **Structs**: Struct definitions, fields, and initialization
- **Methods**: Instance and static methods
- **Builtins**: Built-in functions and methods
- **Modules**: Package system and imports
- **Runtime**: VM and execution model
- **Errors**: Error handling and exceptions
