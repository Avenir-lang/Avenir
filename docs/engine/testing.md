# Testing the Engine

This document describes how to test the compiler and runtime components.

## Test Layout

Tests are organized per package, for example:

- `internal/lexer/*_test.go`
- `internal/parser/*_test.go`
- `internal/types/*_test.go`
- `internal/ir/*_test.go`
- `internal/vm/*_test.go`
- `internal/runtime/builtins/*_test.go`

## Running Tests

From the repo root:

```bash
go test ./...
```

## Parser and Lexer Tests

Typical tests:

1. Lex source into tokens
2. Parse into AST
3. Assert on node structure or error messages

## Type Checker Tests

Use `types.CheckWorldWithBindings` to type‑check a `modules.World`, or parse a
single program and call checker helpers directly. Verify errors and bindings.

## IR and VM Tests

IR tests validate that the compiler emits the expected opcodes (or at least
that execution produces the expected result). VM tests typically compile a
small program, then run `vm.RunMain()` and assert on the result.

## Example: End‑to‑End Test (Conceptual)

```go
src := `
pckg main;
fun main() | void { print("ok"); }
`

// lex/parse
// type check
// compile to IR
// run VM
```

## Tips

- Use small, focused test programs.
- Prefer asserting on returned errors instead of string matching when possible.
- Include tests for runtime errors (e.g., out‑of‑bounds access).
