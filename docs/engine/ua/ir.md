# IR (проміжне представлення)

Цей документ описує байткод IR в Avenir, який використовує VM.
IR визначений у `internal/ir/ir.go`, а генерується в `internal/ir/compiler.go`.

## Огляд

IR — це **стековий байткод** (не SSA). Кожна функція компілюється в `Chunk`,
який містить:

- таблицю констант
- список інструкцій
- кількість локальних слотів

## Основні структури

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

## Категорії інструкцій

- **Стек/локальні**: `OpConst`, `OpLoadLocal`, `OpStoreLocal`, `OpPop`
- **Арифметика**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`, `OpNegate`
- **Порівняння**: `OpEq`, `OpNeq`, `OpLt`, `OpLte`, `OpGt`, `OpGte`
- **Керування потоком**: `OpJump`, `OpJumpIfFalse`
- **Виклики**: `OpCall`, `OpCallValue`, `OpCallBuiltin`, `OpReturn`
- **Дані**: `OpMakeList`, `OpMakeDict`, `OpMakeStruct`, `OpIndex`
- **Виключення**: `OpBeginTry`, `OpEndTry`, `OpThrow`
- **Замикання**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`
- **Асинхронність**: `OpSpawn`, `OpAwait`

Повний список опкодів: `internal/ir/ir.go`.

## Lowering

### Дженеріки

Дженеріки мономорфізуються до/під час збірки IR:

- неінстанційовані generic-декларації пропускаються
- компілятор бере конкретні інстанси з bindings type checker-а

### Async/Await

- Асинхронні функції позначаються прапором `Function.IsAsync`.
- Виклик async-функції компілюється в `OpSpawn`.
- Вираз `await expr` компілюється в `OpAwait`.

`OpSpawn` створює runtime `Future` + `Task` і кладе future у стек.

`OpAwait`:

1. Якщо future готовий — кладе значення в стек.
2. Якщо future завершився з помилкою — пробрасывает/прокидає помилку.
3. Якщо future ще не готовий — призупиняє поточну задачу.

## Посилання

- `internal/ir/ir.go`
- `internal/ir/compiler.go`
