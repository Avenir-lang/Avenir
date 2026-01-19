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

## Notes

- `throw` expects an `error` value; non‑error throws are wrapped as errors.
- `errorMessage(e)` extracts the user‑visible message.

## References

- `internal/lexer/lexer.go`
- `internal/parser/parser.go`
- `internal/types/checker.go`
- `internal/vm/vm.go`

