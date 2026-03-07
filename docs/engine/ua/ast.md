# AST (абстрактне синтаксичне дерево)

Цей документ описує ключові вузли AST у `internal/ast/ast.go`.

## Інтерфейси вузлів

Усі вузли AST реалізують:

```go
type Node interface { Pos() token.Position }
```

Також використовуються маркерні інтерфейси:

- `Stmt` для операторів
- `Expr` для виразів
- `TypeNode` для синтаксису типів

Позиції беруться з токенів і використовуються для діагностики.

## Структура програми

```go
type Program struct {
    Package    *PackageDecl
    Imports    []*ImportDecl
    Funcs      []*FunDecl
    Structs    []*StructDecl
    Interfaces []*InterfaceDecl
}
```

## Декларації

### Функції та методи

`FunDecl` містить:

- `Name`, `Params`, `Return`
- опційний `Receiver` для instance/static методів
- `IsPublic` для `pub fun`

Види receiver:

- `ReceiverInstance`: `(name | Type).method`
- `ReceiverStatic`: `Type.method`

### Структури

`StructDecl` містить:

- `IsPublic`, `IsMutable`
- поля з прапорцями `pub` і `mut`
- compile-time значення за замовчуванням (`DefaultExpr`)

### Інтерфейси

`InterfaceDecl` зберігає лише сигнатури методів (структурна типізація).

## Вузли типів

Ключові варіанти `TypeNode`:

- `SimpleType`, `QualifiedType`
- `ListType`, `DictType`
- `FuncType`, `UnionType`, `OptionalType`

## Вузли операторів

Основні statement-вузли:

- `BlockStmt`, `VarDeclStmt`, `AssignStmt`
- `IfStmt`, `WhileStmt`, `ForStmt`, `ForEachStmt`
- `ReturnStmt`, `TryStmt`, `ThrowStmt`, `BreakStmt`
- `ExprStmt`

## Вузли виразів

Основні expression-вузли:

- `IdentExpr`, `MemberExpr`, `CallExpr`, `IndexExpr`
- `UnaryExpr`, `BinaryExpr`
- літерали (`IntLiteral`, `FloatLiteral`, `StringLiteral`, `BytesLiteral`, `BoolLiteral`)
- `NoneLiteral`, `SomeLiteral`, `ListLiteral`, `DictLiteral`, `StructLiteral`
- `InterpolatedString`, `FuncLiteral`

## Dict і інтерпольовані рядки

Dict-літерали зберігають пари ключ/значення. Ключі можуть бути ідентифікаторами
або рядками:

```avenir
{ name: "Alex", "age": 30 }
```

Інтерпольовані рядки представлені вузлом із масивом частин тексту та виразів.

## Посилання

- `internal/ast/ast.go`

