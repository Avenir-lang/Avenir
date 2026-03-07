# IR (промежуточное представление)

Этот документ описывает байткод IR в Avenir, который использует VM.
IR определён в `internal/ir/ir.go`, а генерируется в `internal/ir/compiler.go`.

## Обзор

IR — это **стековый байткод** (не SSA). Каждая функция компилируется в `Chunk`,
который содержит:

- таблицу констант
- список инструкций
- количество локальных слотов

## Основные структуры

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

## Категории инструкций

- **Стек/локальные**: `OpConst`, `OpLoadLocal`, `OpStoreLocal`, `OpPop`
- **Арифметика**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`, `OpNegate`
- **Сравнения**: `OpEq`, `OpNeq`, `OpLt`, `OpLte`, `OpGt`, `OpGte`
- **Управление потоком**: `OpJump`, `OpJumpIfFalse`
- **Вызовы**: `OpCall`, `OpCallValue`, `OpCallBuiltin`, `OpReturn`
- **Данные**: `OpMakeList`, `OpMakeDict`, `OpMakeStruct`, `OpIndex`
- **Исключения**: `OpBeginTry`, `OpEndTry`, `OpThrow`
- **Замыкания**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`
- **Асинхронность**: `OpSpawn`, `OpAwait`

Полный список опкодов: `internal/ir/ir.go`.

## Lowering

### Дженерики

Дженерики мономорфизуются до/во время сборки IR:

- необобщённые generic-декларации пропускаются
- компилятор берёт конкретные инстансы из bindings type checker-а

### Async/Await

- Асинхронные функции маркируются флагом `Function.IsAsync`.
- Вызов async-функции компилируется в `OpSpawn`.
- Выражение `await expr` компилируется в `OpAwait`.

`OpSpawn` создаёт runtime `Future` + `Task` и помещает future в стек.

`OpAwait`:

1. Если future готов — кладёт значение в стек.
2. Если future завершился с ошибкой — пробрасывает ошибку.
3. Если future ещё не готов — приостанавливает текущую задачу.

## Ссылки

- `internal/ir/ir.go`
- `internal/ir/compiler.go`
