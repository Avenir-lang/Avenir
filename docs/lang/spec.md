# Language Specification (Concise)

This document provides a concise, normative description of Avenir’s core
syntax and semantics. For expanded explanations and examples, see the other
docs in this directory.

## Lexical Structure

- Identifiers: letters/underscore followed by letters/digits/underscore.
- Keywords: `pckg`, `import`, `fun`, `struct`, `pub`, `mut`, `var`, `if`, `else`,
  `while`, `for`, `in`, `return`, `break`, `try`, `catch`, `throw`, `true`,
  `false`, `none`, `some`, `interface`.
- Literals: `int`, `float`, `string` (single or double quoted), `bytes`
  (`b"..."`), `bool`, `none`, `some(...)`.

## Types

- Primitives: `int`, `float`, `string`, `bool`, `bytes`, `void`, `any`, `error`
- Composite: `list<T>`, `dict<T>` (string keys), function types `fun(...) | T`
- Optional: `T?`
- Union: `<T1|T2|...>`
- Struct and interface types

## Expressions

- Arithmetic, comparison, and logical operators.
- String concatenation via `+` is allowed only for `string + string`.
- Indexing: `list[int]`, `bytes[int]`, `dict[string]`.
- Member access: `expr.field` and `expr.method(...)`.

## Statements

- Variable declarations: `var name | Type = expr;`
- Assignment: `name = expr;`
- `if`, `while`, `for`, `for (item in list)` loops.
- `return`, `break`, `throw`, `try/catch`.

## Modules

- Each file declares a package with `pckg`.
- Imports use dotted module paths: `import std.net as net;`.
- Folder-based modules resolve `path/path.av` with file-to-struct mapping when
  structs are present.

## Runtime Semantics

- All runtime errors (including builtins) are thrown as `error` values and are
  catchable via `try / catch`.
- Built-in functions and methods are part of the language core and are invoked
  directly by the VM.

For detailed syntax, builtins, and standard library APIs, see:
`syntax.md`, `types.md`, `operators.md`, `builtins.md`, and `docs/std/`.