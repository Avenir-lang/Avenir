# AST (Abstract Syntax Tree)

This document describes the core AST node types in `internal/ast/ast.go`.

## Node Interfaces

All AST nodes implement:

```go
type Node interface { Pos() token.Position }
```

And the marker interfaces:

- `Stmt` for statements
- `Expr` for expressions
- `TypeNode` for type syntax

Positions are derived from tokens and used for diagnostics.

## Program Structure

```go
type Program struct {
    Package    *PackageDecl
    Imports    []*ImportDecl
    Funcs      []*FunDecl
    Structs    []*StructDecl
    Interfaces []*InterfaceDecl
}
```

## Declarations

### Functions and Methods

`FunDecl` contains:

- `Name`, `Params`, `Return`
- Optional `Receiver` for instance/static methods
- `IsPublic` for `pub fun`

Receiver kinds:

- `ReceiverInstance`: `(name | Type).method`
- `ReceiverStatic`: `Type.method`

### Structs

`StructDecl` includes:

- `IsPublic` and `IsMutable`
- Fields with per‑field `pub` and `mut`
- Optional compile‑time defaults (`DefaultExpr`)

### Interfaces

`InterfaceDecl` stores method signatures only (structural typing).

## Type Nodes

Key `TypeNode` variants:

- `SimpleType` (builtins and local names)
- `QualifiedType` (`net.Socket`)
- `ListType` (`list<T1, T2>`)
- `DictType` (`dict<T>`)
- `FuncType` (`fun(...) | T`)
- `UnionType` (`<A|B|...>`)
- `OptionalType` (`T?`)

## Statement Nodes

Important statement nodes:

- `BlockStmt`
- `VarDeclStmt`
- `AssignStmt`
- `IfStmt` / `Else`
- `WhileStmt`
- `ForStmt`
- `ForEachStmt`
- `ReturnStmt`
- `TryStmt`
- `ThrowStmt`
- `BreakStmt`
- `ExprStmt`

## Expression Nodes

Common expression nodes:

- `IdentExpr`
- `MemberExpr` and `CallExpr`
- `IndexExpr`
- `UnaryExpr` / `BinaryExpr`
- Literal nodes: `IntLiteral`, `FloatLiteral`, `StringLiteral`, `BytesLiteral`, `BoolLiteral`
- `NoneLiteral`, `SomeLiteral`
- `ListLiteral`, `DictLiteral`, `StructLiteral`
- `InterpolatedString` with `StringTextPart` and `StringExprPart`
- `FuncLiteral`

## Dict and Interpolated Strings

Dict literals store entries as key/value expressions, where keys can be
identifier tokens (converted to strings) or string literals:

```avenir
{ name: "Alex", "age": 30 }
```

Interpolated strings are represented as:

```go
type InterpolatedString struct {
    Parts []StringPart // StringTextPart | StringExprPart
}
```

The parser produces a single `InterpolatedString` node for a literal that
contains `${...}` segments.

## Example

Source:

```avenir
var x | int = 1 + 2;
```

AST (conceptual):

```
VarDeclStmt(name="x", type="int",
  value=BinaryExpr("+", IntLiteral(1), IntLiteral(2)))
```

## Notes

- AST nodes are syntax‑level only; type information is tracked in the checker
  and binding tables, not embedded in the AST.
- `Pos()` is used heavily for error reporting in the parser and checker.

## References

- `internal/ast/ast.go`

