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
- **Control flow**: `OpJump`, `OpJumpIfFalse`
- **Calls**: `OpCall`, `OpCallValue`, `OpCallBuiltin`, `OpReturn`
- **Data**: `OpMakeList`, `OpMakeDict`, `OpMakeStruct`, `OpIndex`
- **Fields**: `OpLoadField`, `OpStoreField`
- **Strings**: `OpStringify`, `OpConcatString`
- **Optionals**: `OpMakeSome`
- **Exceptions**: `OpBeginTry`, `OpEndTry`, `OpThrow`
- **Closures**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`

See `internal/ir/ir.go` for the full opcode list.

## Lowering Highlights

### List and Dict Literals

List and dict literals compile to `OpMakeList` and `OpMakeDict` with their
elements pushed first.

### Struct Literals

Struct literals compile to `OpMakeStruct` with fields pushed in declared order.

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
