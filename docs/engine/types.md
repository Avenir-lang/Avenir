# Types and Type Checker

This document describes the static type system and the checker in
`internal/types`.

## Overview

Type checking is explicit and static: all variables and parameters must have
declared types. The checker validates program structure, resolves names, and
produces binding metadata used by the IR compiler.

Key files:

- `internal/types/types.go` (type definitions)
- `internal/types/checker.go` (type checker)

## Type Kinds

### Basic Types

`int`, `float`, `string`, `bool`, `bytes`, `void`, `any`, `error`.

### List Types

`list<T>` supports **multiple element types**:

- `list<int>`
- `list<int, string>`

Internally, `List.ElementTypes` is a list of valid element types. List literals
collect unique element types in order of appearance.

### Dict Types

`dict<T>` maps **string keys** to values of type `T`. Dict literals infer `T`
as the union of value types if needed.

### Optional Types

`T?` represents an optional. `none` is the only null‑like value, and
`some(expr)` wraps a value into an optional.

Example:

```avenir
var n | int? = some(42);
var m | int? = none;
```

### Union Types

`<A|B|C>` allows a value to be any of the variants. Unions are order‑insensitive
for equality.

Example:

```avenir
var mixed | <int|string> = 1;
```

### Structs and Interfaces

- Structs are **nominal**: names define identity.
- Interfaces are **structural**: a type satisfies an interface if it provides
  all required methods with matching signatures.

### Function Types

`fun(T1, T2) | R` for first‑class functions. Parameter and return types are
structural for equivalence.

## Name Resolution

The checker populates scopes with:

- Builtins (functions only)
- Local declarations
- Imported module symbols (by alias and full path)

Qualified types (e.g. `net.Socket`) are resolved by inserting the full module
path into scope during import processing.

## Checker Flow

High‑level flow (simplified):

1. Declare builtins from `internal/runtime/builtins`.
2. Load modules and merge multi‑file programs.
3. Declare types (structs, interfaces).
4. Declare functions and methods.
5. Type‑check all statements and expressions.
6. Produce `Bindings` metadata for IR compilation.

## Assignability Rules (Highlights)

- `any` accepts all types.
- A value of type `T` is assignable to `T?` (wrapped as `some` by the runtime).
- `none` is assignable to any optional type.
- Unions allow assignment if the source type matches any variant.
- Lists are assignable if each element type is assignable to some target
  element type.
- Dicts are assignable if the value type is assignable to the target value
  type.
- Structs require identical names (nominal).
- Interfaces require a full method match (structural).

Note: the compiler only emits `OpMakeSome` for explicit `some(...)` literals.
Optional promotion is a type rule; there is no implicit runtime wrapping beyond
explicit `some`.

## Operators

The checker enforces operator rules, for example:

- `+` supports numeric addition and `string + string` only.
- `-`, `*`, `/`, `%` require numeric operands.
- Comparisons allow numeric operands and `==`/`!=` allow broader comparisons.

## Builtins and Methods

Builtins are registered via `builtins.Meta` and injected into scope as regular
functions (`ReceiverType == TypeVoid`). Builtin methods are resolved at member
access time based on the receiver’s type.

## Errors

Type errors are collected and returned as a slice of `types.Error` with source
positions.

## References

- `internal/types/types.go`
- `internal/types/checker.go`

