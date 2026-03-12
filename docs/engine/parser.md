# Parser

This document describes Avenir’s recursive‑descent parser, implemented in
`internal/parser/parser.go`.

## Overview

Parsing is a single‑pass recursive descent over a lookahead token stream. The
parser builds AST nodes from `internal/ast` and accumulates errors instead of
panicking.

## Program Structure

At the top level the parser expects:

1. `pckg` declaration
2. `import` declarations
3. `struct` declarations
4. `interface` declarations
5. `fun` declarations (including methods)

The parser enforces package declarations and reports errors for missing or
malformed constructs.

## Expression Precedence

Expressions are parsed by precedence‑layer functions:

1. `parseOr` → `||`
2. `parseAnd` → `&&`
3. `parseEquality` → `==`, `!=`
4. `parseRelational` → `<`, `<=`, `>`, `>=`
5. `parseAdditive` → `+`, `-`
6. `parseMultiplicative` → `*`, `/`, `%`
7. `parseUnary` → `!`, unary `-`, `await`
8. `parsePostfix` → member access, calls, indexing
9. `parsePrimary` → literals, identifiers, grouped expressions

## Statements

Supported statements include:

- Variable declarations: `var name | Type = expr;`
- Assignments: `name = expr;`
- Struct field assignments: `obj.field = expr;` and `obj.nested.field = expr;`
- Expression statements
- `if` / `else`
- `while`
- C‑style `for`
- `for (item in list)` foreach
- `return`

## Member Access

The parser supports chained member access in both expressions and assignments:

```avenir
// Expression chaining
var result | string = obj.a.b.c();

// Field assignment chaining
obj.nested.field = value;
```

Chained access is handled in the statement postfix loop with a `Dot` case that
continues parsing member expressions until a non-dot token is encountered.

Additional statements:
- `try` / `catch`
- `throw`
- `break`
- `continue`
- `switch` / `case` / `default`
- `defer` (call expression only)

## Async Syntax Parsing

### Async Function Declarations

`parseFunDecl` accepts optional modifiers in this order:

1. `pub` (optional)
2. `async` (optional)
3. `fun` (required)

The parser stores async marker in `ast.FunDecl.IsAsync`.

### Await Expressions

`await` is parsed in `parseUnary`, which gives it unary-operator precedence.

Conceptually:

```avenir
await a + b
```

is parsed as:

```avenir
(await a) + b
```

The parser emits `ast.AwaitExpr` and leaves type validation to the type checker.

### `if` condition sugar

The parser supports:

```avenir
if (a > 0; a < 10; flag) { ... }
```

Semicolons are desugared into `&&` at parse time.

## Calls and Member Access

`parsePostfix` handles:

- Member access: `expr.name`
- Optional member access: `expr?.name`
- Calls: `expr(args...)`
- Optional calls: `expr?.(args...)`, `expr?.method(args...)`
- Indexing: `expr[index]`

Named arguments use `name = expr` syntax. The parser only recognizes them in
call argument lists.

## Literals

Primary literals include:

- `int`, `float`, `string`, `bytes`, `bool`
- `none`, `some(expr)`
- List literal: `[a, b, c]`
- Dict literal: `{ key: value, "key": value }`
- Struct literal: `TypeName{ field = value }`
- Interpolated string: `"x=${expr}"`

## Types

Type parsing supports:

- Simple types: `int`, `float`, `string`, `bool`, `bytes`, `void`, `any`, `error`
- Qualified types: `net.Socket`
- List types: `list<T1, T2>`
- Dict types: `dict<K, V>` or `dict<V>` (shorthand for `dict<string, V>`)
- Optional types: `T?`
- Union types: `<T1|T2|...>`
- Function types: `fun(T1, T2) | R`

The `?` suffix is allowed on simple, qualified, list, dict, and union types.

## Error Handling

The parser collects errors via `p.errorf(...)` and continues when possible.
`p.Errors()` returns all parse errors.

## References

- `internal/parser/parser.go`
- `internal/ast/ast.go`
