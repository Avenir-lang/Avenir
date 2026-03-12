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
- `Future`
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
- Deferred call stack (`DeferStack`)

### Call Convention

Arguments are pushed by the caller. The callee’s base is `sp - numArgs`.
`NumLocals` determines how many slots to reserve for locals beyond parameters.

## Execution Loop

The VM fetches instructions from the current frame and executes them in a loop.
Key behaviors:

- `OpConst` pushes constants onto the stack
- `OpLoadLocal`/`OpStoreLocal` access frame‑local slots
- Arithmetic and comparisons pop operands and push results
- `OpJumpIfNone` handles optional-chain branching
- `OpCall`/`OpCallValue` call functions or closures
- `OpCallBuiltin` routes to runtime builtins
- `OpPushDefer` stores deferred calls for execution at return time
- `OpSpawn` wraps async call result into `Future`
- `OpAwait` reads or suspends on `Future`

### Optional Chaining Runtime Semantics

`OpJumpIfNone` inspects the top stack value:

- If it is `none`, execution jumps to the provided target.
- If it is `some(v)`, the VM unwraps it in-place to `v` on the stack.

This supports optional chaining short-circuiting while allowing the non-`none`
path to operate on the inner value.

### Deferred Calls on Return

`OpPushDefer` captures `(callee, args...)` and pushes the deferred call onto the
current frame's `DeferStack`.

On `OpReturn`, deferred calls are executed in LIFO order before the frame is
popped.

## Async Execution Model

### Async Main Entry

`RunMain` checks `main` function metadata:

- sync `main` → direct `callClosure`
- async `main` (`Function.IsAsync`) → `runAsyncMain`

`runAsyncMain` initializes runtime scheduler, creates `Future` for main result,
wraps main execution into a `Task`, schedules it, and starts event loop.

### `OpSpawn`

`OpSpawn` uses function index + argument count from IR instruction.

Current behavior:

1. Execute target closure immediately through `callClosure`.
2. Create `runtime.Future`.
3. Resolve or reject that future with call result/error.
4. Push `Future` value to VM stack.

### `OpAwait`

`OpAwait` pops a value and expects `Future` kind.

- Ready + success: pushes resolved result.
- Ready + error: propagates as throwable error.
- Not ready:
  - in async task context: registers waiter task, snapshots VM state
    (stack/frames/handlers), marks task suspended, returns suspension sentinel;
  - outside async task context: runtime error (`future not ready in non-async context`).

When the awaited future resolves/rejects, waiter tasks are rescheduled by the
runtime scheduler/event loop.

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

Struct values thrown via `OpThrow` are passed through to catch handlers without
wrapping, enabling typed catch clauses to match on struct type.

`OpIsStructType` peeks TOS and pushes `true`/`false` indicating whether the
value is a struct matching the given type index. This is used by the typed catch
clause dispatch chain.

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

