# AST (абстрактное синтаксическое дерево)

Этот документ описывает ключевые узлы AST в `internal/ast/ast.go`.

## Интерфейсы узлов

Все узлы AST реализуют:

```go
type Node interface { Pos() token.Position }
```

Также используются маркерные интерфейсы:

- `Stmt` для операторов
- `Expr` для выражений
- `TypeNode` для синтаксиса типов

Позиции берутся из токенов и используются в диагностике.

## Структура программы

```go
type Program struct {
    Package    *PackageDecl
    Imports    []*ImportDecl
    Funcs      []*FunDecl
    Structs    []*StructDecl
    Interfaces []*InterfaceDecl
}
```

## Декларации

### Функции и методы

`FunDecl` включает:

- `Name`, `Params`, `Return`
- опциональный `Receiver` для instance/static методов
- `IsPublic` для `pub fun`

Виды receiver:

- `ReceiverInstance`: `(name | Type).method`
- `ReceiverStatic`: `Type.method`

### Структуры

`StructDecl` содержит:

- `IsPublic`, `IsMutable`
- поля с флагами `pub` и `mut`
- compile-time значения по умолчанию (`DefaultExpr`)

### Интерфейсы

`InterfaceDecl` хранит только сигнатуры методов (структурная типизация).

## Узлы типов

Ключевые варианты `TypeNode`:

- `SimpleType`, `QualifiedType`
- `ListType`, `DictType`
- `FuncType`, `UnionType`, `OptionalType`

## Узлы операторов

Основные statement-узлы:

- `BlockStmt`, `VarDeclStmt`, `AssignStmt`
- `IfStmt`, `WhileStmt`, `ForStmt`, `ForEachStmt`
- `ReturnStmt`, `TryStmt`, `ThrowStmt`, `BreakStmt`
- `ExprStmt`

## Узлы выражений

Основные expression-узлы:

- `IdentExpr`, `MemberExpr`, `CallExpr`, `IndexExpr`
- `UnaryExpr`, `BinaryExpr`
- литералы (`IntLiteral`, `FloatLiteral`, `StringLiteral`, `BytesLiteral`, `BoolLiteral`)
- `NoneLiteral`, `SomeLiteral`, `ListLiteral`, `DictLiteral`, `StructLiteral`
- `InterpolatedString`, `FuncLiteral`

## Dict и интерполированные строки

Dict-литералы хранят пары ключ/значение. Ключи могут быть идентификаторами
или строками:

```avenir
{ name: "Alex", "age": 30 }
```

Интерполированные строки представлены узлом с массивом частей текста и
выражений.

## Ссылки

- `internal/ast/ast.go`

