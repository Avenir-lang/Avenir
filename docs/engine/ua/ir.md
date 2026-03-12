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
- **Виключення**: `OpBeginTry`, `OpEndTry`, `OpThrow`, `OpIsStructType`
- **Замикання**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`
- **Асинхронність**: `OpSpawn`, `OpAwait`

Повний список опкодів: `internal/ir/ir.go`.

## Lowering

### Дженеріки

Дженеріки мономорфізуються до/під час збірки IR:

- неінстанційовані generic-декларації пропускаються
- компілятор бере конкретні інстанси з bindings type checker-а

### Try/Catch

`try { ... } catch (...) { ... }` компілюється в:

1. `OpBeginTry` з IP обробника
2. інструкції try-блоку
3. `OpEndTry`
4. блок обробника

При наявності кількох типізованих catch-клауз блок обробника генерує
ланцюжок перевірок:

1. Для кожного `catch (var | StructType)`:
   - `OpIsStructType A` (A = індекс типу структури) — peek TOS, push bool
   - `OpJumpIfFalse` до наступної клаузи
   - збереження значення в локальну змінну, тіло клаузи, `OpJump` в кінець
2. Для `catch (var | error)`: fallback-клауза
3. Фінальний `OpThrow` перекидає необроблені помилки

### Async/Await

- Асинхронні функції позначаються прапором `Function.IsAsync`.
- Виклик async-функції компілюється в `OpSpawn`.
- Вираз `await expr` компілюється в `OpAwait`.

`OpSpawn` і `OpAwait` — async-границі на рівні VM:

1. `OpSpawn` споживає аргументи виклику і кладе в стек `Future`.
2. `OpAwait` споживає `Future`:
   - ready + success: кладе результат у стек
   - ready + error: прокидає/викидає помилку
   - not ready: призупиняє поточний async-контекст задачі

У поточній реалізації `OpSpawn` виконує цільове замикання одразу та
обгортає результат/помилку в `Future`; механізм suspend/resume забезпечується
через `OpAwait` і runtime scheduler/event loop.

## Посилання

- `internal/ir/ir.go`
- `internal/ir/compiler.go`
