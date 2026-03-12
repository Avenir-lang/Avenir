# Errors and Diagnostics

This document describes error handling across the compiler and runtime.

## Compile‑Time Errors

### Lexer Errors

The lexer records invalid tokens (unterminated strings, invalid escapes, etc.)
with line/column positions. Errors are retrieved via `lexer.Errors()`.

### Parser Errors

The parser collects syntax errors via `p.errorf(...)`. Parsing continues when
possible to report multiple errors in a single run.

### Type Checker Errors

The type checker returns `types.Error` entries that include positions and
messages. Examples include:

- Type mismatches
- Undefined names
- Invalid assignments
- Interface conformance errors

## Runtime Errors

At runtime, all errors are represented as `error` values and are throwable:

- Division by zero
- Index out of bounds
- Invalid builtin usage
- Failed I/O or network operations

## Try/Catch Semantics

The compiler emits:

- `OpBeginTry` with a handler IP
- `OpEndTry` to pop the handler
- `OpThrow` to throw a value

The VM installs handlers on a stack. When an error is raised, it:

1. Unwinds to the nearest handler
2. Restores stack depth
3. Pushes the error value
4. Jumps to the handler IP

If no handler exists, the error propagates to the VM entry point and terminates
execution.

### Typed Catch Clauses

When a `try` block has multiple typed catch clauses, the compiler emits a chain
of type checks at the handler entry point:

1. `OpIsStructType` peeks the thrown value and pushes `true`/`false` indicating
   whether it matches the target struct type index.
2. `OpJumpIfFalse` skips to the next clause if the type does not match.
3. On match, the thrown value is stored in a local and the clause body executes,
   then `OpJump` skips past remaining clauses.
4. After all clauses, an `OpThrow` re‑throws unmatched errors.

A `catch (e | error)` clause matches any error and acts as a fallback.

### Throws Declarations

Functions can declare thrown error types using `!` after the return type:

```avenir
fun f() | void ! FileNotFound, PermissionDenied { ... }
```

The type checker validates that `throw` expressions match declared throws. The
`Func` type in `types.go` carries a `Throws []Type` field.

## Notes

- `throw` accepts `error` values and struct values.
- Struct values thrown as errors are passed through to catch handlers without
  wrapping.
- `errorMessage(e)` extracts the user‑visible message from `error` values.

## References

- `internal/lexer/lexer.go`
- `internal/parser/parser.go`
- `internal/types/checker.go`
- `internal/vm/vm.go`

