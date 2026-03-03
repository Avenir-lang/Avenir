# Virtual Machine

This document describes the Avenir VM in `internal/vm/vm.go`.

## Overview

The VM is a stack‑based interpreter for the IR bytecode. It executes functions
compiled by `internal/ir` and uses a runtime environment for builtins and host
services.

## Value Representation

Runtime values are stored as `value.Value` (see `internal/value`), with a
`Kind` discriminant, such as:

- `Int`, `Float`, `String`, `Bool`, `Bytes`
- `List`, `Dict`
- `Struct`
- `Optional` (`some` / `none`)
- `Closure`
- `Error`

`Value.String()` is used for stringification and printing.

## Stack and Frames

The VM maintains:

- A value stack (`stack` + `sp`)
- A call stack of `Frame` objects

Each frame tracks:

- The current closure/function
- Instruction pointer (`IP`)
- Base stack index for locals

### Call Convention

Arguments are pushed by the caller. The callee’s base is `sp - numArgs`.
`NumLocals` determines how many slots to reserve for locals beyond parameters.

## Execution Loop

The VM fetches instructions from the current frame and executes them in a loop.
Key behaviors:

- `OpConst` pushes constants onto the stack
- `OpLoadLocal`/`OpStoreLocal` access frame‑local slots
- Arithmetic and comparisons pop operands and push results
- `OpCall`/`OpCallValue` call functions or closures
- `OpCallBuiltin` routes to runtime builtins

## Builtins Dispatch

`OpCallBuiltin` calls `runtime.CallBuiltin`, which performs:

1. ID lookup in the builtin registry
2. Argument conversion to `[]interface{}`
3. Builtin invocation with `runtime.Env`
4. Conversion of the result back to `value.Value`

## Errors and Exceptions

Runtime errors are converted to `error` values and thrown:

- `raiseError(err)` wraps the error into `value.ErrorValue`
- `throwValue` unwinds to the nearest handler installed by `OpBeginTry`

Thrown non‑error values are wrapped as `error` with the message
`"thrown non-error: ..."` for consistency.

## Closures and Upvalues

Closures capture outer variables:

- `OpClosure` creates a closure with upvalues
- `OpLoadUpvalue` / `OpStoreUpvalue` read/write captured values
- Open upvalues are closed when their owning frame returns

## Notes and Pitfalls

- Indexing errors (out of range, missing dict key) raise runtime errors.
- String concatenation uses a `strings.Builder` for efficiency.
- Lists are not mutated by list methods; dicts can be mutated via `dict.set`.

## References

- `internal/vm/vm.go`
- `internal/value/value.go`

