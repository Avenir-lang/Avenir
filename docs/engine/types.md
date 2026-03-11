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

### Future Types (Async)

The checker models async return values with `Future<T>`:

- `async fun ... | T` is inserted into scope as `fun(...) | Future<T>`.
- Inside async function bodies, `return expr;` is still checked against `T`
  (the inner type), not `Future<T>`.
- `await expr` requires `expr` to have type `Future<X>` and produces `X`.

If `await` receives a non-future expression, checker reports an error.

### Generics

The checker supports explicit generics for user-defined structs and functions:

- Generic struct declarations: `struct Box<T> { ... }`
- Generic function declarations: `fun identity<T>(x | T) | T { ... }`
- Generic type usage: `Box<int>`
- Generic call usage: `identity<int>(10)`

Type arguments are explicit. Generic type argument inference is not currently
implemented.

Built-in collections (`list<...>`, `dict<...>`) are parametric built-ins and are
handled separately from user-defined generics.

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

For generics, declaration and checking include additional steps:

1. Register generic templates as `GenericStruct` / `GenericFunc` symbols.
2. On `Name<TypeArgs>` use, resolve concrete type arguments.
3. Instantiate concrete types/signatures via substitution.
4. Register monomorphized structs/functions under generated names
   (for example, `Box$int`, `identity$int`).
5. Emit these monomorphized entries through `Bindings` so IR compilation can
   include only instantiated versions.

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

Note: the compiler emits `OpMakeSome` for explicit `some(...)` literals and for
successful optional-chaining non-none paths (`?.`).

Optional promotion remains a type rule; plain assignment to `T?` does not imply
general-purpose runtime wrapping for arbitrary expressions.

## Operators

The checker enforces operator rules, for example:

- `+` supports numeric addition and `string + string` only.
- `-`, `*`, `/`, `%` require numeric operands.
- Comparisons allow numeric operands and `==`/`!=` allow broader comparisons.

## Builtins and Methods

Builtins are registered via `builtins.Meta` and injected into scope as regular
functions (`ReceiverType == TypeVoid`). Builtin methods are resolved at member
access time based on the receiver’s type.

## Decorators

Decorators use expression-based syntax (`@<expr>`) and are applied at init-time.

The checker validates decorators on function declarations via `checkDecorators`:

1. Type-check the decorator expression via `checkExpr(dec.Expr)`.
2. If the result is a `*Func`, verify it accepts exactly one function parameter
   matching the decorated function's type and returns the same function type
   (`validateDecoratorFunc`).
3. If the result is a `*GenericFunc`, infer type arguments from the decorated
   function's signature via `inferDecoratorTypeArgs`, instantiate, verify, and
   create a synthetic `IdentExpr` with the monomorphized name.

Since decorators are arbitrary expressions, parameterized decorators like
`@cache(60)` are handled naturally: the expression `cache(60)` is type-checked
as a call expression, and its result type is validated as a decorator function.

For instance methods, the decorated function type includes the receiver as the
first parameter (for example, `fun(Point, int) | int`), so the decorator must
accept and return that receiver-inclusive function type.

Resolved decorator info is stored in `Bindings.Decorators` as `[]*DecoratorInfo`,
keyed by `*ast.FunDecl`. Each `DecoratorInfo` contains the decorator expression
(`ast.Expr`) and the resolved function type.

The IR compiler generates a synthetic `__init__` function that applies decorators
at module initialization using `OpSetFunc`. For each decorated function, the init
function pushes the original as a closure, compiles the decorator expression, calls
it, and emits `OpSetFunc` to replace the function slot. The VM stores the resulting
closure in `closureOverrides` so that `OpCall`, `OpClosure`, and `OpSpawn` use the
decorated version. This is a one-time operation with zero per-call overhead.

## Variadic Generics (TypePack)

The type system supports variadic type parameters via `TypePack`:

- `TypePack` holds a `[]Type` representing a pack of concrete types.
- `SubstituteType` expands `TypePack` in `Func` parameter lists: if a substituted
  parameter type is a `TypePack`, its types are spliced into the parameter list.
- `MonomorphKey` flattens `TypePack` entries into the key string.
- `typeOfTypeNode` handles `*ast.TypePackExpansion` by returning a `TypeVar` that
  will be expanded during substitution.

## Bindings Output

`Bindings` includes expression/member resolution, generic instantiation, and decorator data:

- `Idents`, `Members`, `ExprTypes`
- `MonomorphizedStructs` (`monoName -> *types.Struct`)
- `MonomorphizedFuncs` (`monoName -> *ast.FunDecl`)
- `Decorators` (`*ast.FunDecl -> []*DecoratorInfo`)

The IR compiler uses monomorphized maps to collect concrete generic instances and
decorator maps to generate the `__init__` function for init-time decorator application.

## Errors

Type errors are collected and returned as a slice of `types.Error` with source
positions.

## References

- `internal/types/types.go`
- `internal/types/checker.go`

