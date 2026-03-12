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
- **Исключения**: `OpBeginTry`, `OpEndTry`, `OpThrow`, `OpIsStructType`
- **Замыкания**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`
- **Асинхронность**: `OpSpawn`, `OpAwait`

Полный список опкодов: `internal/ir/ir.go`.

## Lowering

### Дженерики

Дженерики мономорфизуются до/во время сборки IR:

- необобщённые generic-декларации пропускаются
- компилятор берёт конкретные инстансы из bindings type checker-а

### Try/Catch

`try { ... } catch (...) { ... }` компилируется в:

1. `OpBeginTry` с IP обработчика
2. инструкции try-блока
3. `OpEndTry`
4. блок обработчика

При наличии нескольких типизированных catch-клауз блок обработчика генерирует
цепочку проверок:

1. Для каждого `catch (var | StructType)`:
   - `OpIsStructType A` (A = индекс типа структуры) — peek TOS, push bool
   - `OpJumpIfFalse` к следующей клаузе
   - сохранение значения в локальную переменную, тело клаузы, `OpJump` в конец
2. Для `catch (var | error)`: fallback-клауза
3. Финальный `OpThrow` перебрасывает необработанные ошибки

### Async/Await

- Асинхронные функции маркируются флагом `Function.IsAsync`.
- Вызов async-функции компилируется в `OpSpawn`.
- Выражение `await expr` компилируется в `OpAwait`.

`OpSpawn` и `OpAwait` — границы async на уровне VM:

1. `OpSpawn` потребляет аргументы вызова и кладёт в стек `Future`.
2. `OpAwait` потребляет `Future`:
   - ready + success: кладёт результат в стек
   - ready + error: пробрасывает/кидает ошибку
   - not ready: приостанавливает текущий async-контекст задачи

В текущей реализации `OpSpawn` выполняет целевое замыкание сразу и
оборачивает результат/ошибку в `Future`; механизм suspend/resume обеспечивается
через `OpAwait` и runtime scheduler/event loop.

## Ссылки

- `internal/ir/ir.go`
- `internal/ir/compiler.go`
