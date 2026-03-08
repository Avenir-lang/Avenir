# IR (Intermediate Representation)

This document describes Avenir’s bytecode IR used by the VM. It is defined in
`internal/ir/ir.go` and produced by the compiler in `internal/ir/compiler.go`.

## Overview

The IR is a **stack‑based bytecode**, not SSA. Each function compiles to a
`Chunk` containing:

- A constant table
- A list of instructions
- A local‑slot count

## Core Structures

```go
type Module struct {
    Functions   []*Function
    StructTypes []StructTypeInfo
    MainIndex   int
}
```

```go
type Function struct {
    Name      string
    NumParams int
    Chunk     Chunk
    Upvalues  []UpvalueInfo
    IsAsync   bool
}
```

```go
type Chunk struct {
    Code      []Instruction
    Consts    []Constant
    NumLocals int
}
```

## Constants

Constants are stored in a table and referenced by index:

- `ConstInt`, `ConstFloat`
- `ConstString`, `ConstBool`
- `ConstBytes`, `ConstNone`

## Instruction Categories

Key instruction categories include:

- **Stack/locals**: `OpConst`, `OpLoadLocal`, `OpStoreLocal`, `OpPop`
- **Arithmetic**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`, `OpNegate`
- **Comparisons**: `OpEq`, `OpNeq`, `OpLt`, `OpLte`, `OpGt`, `OpGte`
- **Control flow**: `OpJump`, `OpJumpIfFalse`, `OpJumpIfNone`
- **Calls**: `OpCall`, `OpCallValue`, `OpCallBuiltin`, `OpPushDefer`, `OpReturn`
- **Data**: `OpMakeList`, `OpMakeDict`, `OpMakeStruct`, `OpIndex`
- **Fields**: `OpLoadField`, `OpStoreField`
- **Strings**: `OpStringify`, `OpConcatString`
- **Optionals**: `OpMakeSome`
- **Exceptions**: `OpBeginTry`, `OpEndTry`, `OpThrow`
- **Closures**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`
- **Async**: `OpSpawn`, `OpAwait`

See `internal/ir/ir.go` for the full opcode list.

## Lowering Highlights

### Generic Functions and Structs

Generics are monomorphized before or during IR collection:

- Uninstantiated generic declarations are skipped.
- The compiler consumes monomorphized entries from type-checker bindings.
- Each concrete instantiation gets its own function/type name
  (for example, `identity$int`, `Box$int`).

Only concrete instantiations referenced by the program are emitted.

### List and Dict Literals

List and dict literals compile to `OpMakeList` and `OpMakeDict` with their
elements pushed first.

### Struct Literals

Struct literals compile to `OpMakeStruct` with fields pushed in declared order.

For generic struct literals, the compiler resolves the monomorphized struct name
and emits `OpMakeStruct` for that concrete struct type index.

### Interpolated Strings

`"x=${expr}"` lowers to:

1. Compile each part
2. `OpStringify` for expression parts
3. `OpConcatString` between parts

### Try/Catch

`try { ... } catch (...) { ... }` compiles to:

1. `OpBeginTry` with handler IP
2. try‑block instructions
3. `OpEndTry`
4. handler block

### Switch / Continue

- `switch` lowers to equality checks (`OpEq`) plus conditional jumps
  (`OpJumpIfFalse`) for each `case`.
- Each matched case body ends with `OpJump` to skip the remaining clauses.
- `continue` lowers to a jump back to the loop-specific continue target.

### Optional Chaining

Optional chains (`?.`) use `OpJumpIfNone`:

1. Evaluate receiver/callee.
2. `OpJumpIfNone` jumps to none-path when the value is `none`.
3. Non-none path evaluates member/call and wraps the result via `OpMakeSome`.

This keeps optional chain expression results in optional form.

### Defer

`defer` lowers to `OpPushDefer`:

1. Evaluate and capture deferred call arguments.
2. Evaluate deferred callee.
3. Emit `OpPushDefer` with captured argument count.

Deferred calls execute later in VM return handling (LIFO order).

### Async/Await Lowering

Async metadata and opcodes are emitted as follows:

- `ast.FunDecl.IsAsync` is propagated to `ir.Function.IsAsync`.
- Direct calls to async function declarations lower to `OpSpawn`.
- `await expr` lowers to `OpAwait` after compiling `expr`.

`OpSpawn` and `OpAwait` are VM-level async boundary instructions:

1. `OpSpawn` consumes compiled call arguments and pushes a `Future` value.
2. `OpAwait` consumes a `Future`:
   - ready + success: pushes resolved value
   - ready + failure: throws/propagates error
   - not ready: suspends current async task context

In the current implementation, `OpSpawn` executes the callee closure
immediately and wraps completion/error into a `Future`; suspension/resume is
driven by `OpAwait` + scheduler/event-loop task coordination.

## Example

Source:

```avenir
print("a" + "b");
```

Lowered IR (conceptual):

```
OpConst "a"
OpConst "b"
OpConcatString
OpCallBuiltin print, 1
```

## References

- `internal/ir/ir.go`
- `internal/ir/compiler.go`
