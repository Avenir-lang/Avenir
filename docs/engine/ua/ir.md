# IR (проміжне представлення)

Цей документ описує байткод IR в Avenir, який використовує VM. IR визначений у `internal/ir/ir.go`, а генерується компілятором у `internal/ir/compiler.go`.

## Огляд

IR — це **стековий байткод** (не SSA). Кожна функція компілюється в `Chunk`, який містить:

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

## Константи

Константи зберігаються в таблиці і адресуються за індексом:

- `ConstInt`, `ConstFloat`
- `ConstString`, `ConstBool`
- `ConstBytes`, `ConstNone`

## Категорії інструкцій

Основні категорії інструкцій:

- **Стек/локальні**: `OpConst`, `OpLoadLocal`, `OpStoreLocal`, `OpPop`
- **Арифметика**: `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod`, `OpNegate`
- **Порівняння**: `OpEq`, `OpNeq`, `OpLt`, `OpLte`, `OpGt`, `OpGte`
- **Керування потоком**: `OpJump`, `OpJumpIfFalse`, `OpJumpIfNone`
- **Виклики**: `OpCall`, `OpCallValue`, `OpCallBuiltin`, `OpPushDefer`, `OpReturn`
- **Дані**: `OpMakeList`, `OpMakeDict`, `OpMakeStruct`, `OpIndex`
- **Поля**: `OpLoadField`, `OpStoreField`
- **Рядки**: `OpStringify`, `OpConcatString`
- **Optional**: `OpMakeSome`
- **Виключення**: `OpBeginTry`, `OpEndTry`, `OpThrow`, `OpIsStructType`
- **Замикання**: `OpClosure`, `OpLoadUpvalue`, `OpStoreUpvalue`
- **Асинхронність**: `OpSpawn`, `OpAwait`

Повний список опкодів: `internal/ir/ir.go`.

## Lowering

### Узагальнені функції та структури

Дженеріки мономорфізуються до або під час збірки IR:

- Неінстанційовані generic-декларації пропускаються.
- Компілятор споживає мономорфізовані записи з bindings type checker-а.
- Кожна конкретна інстанціація отримує власне ім'я функції/типу (наприклад, `identity$int`, `Box$int`).

Генеруються лише ті конкретні інстанціації, на які є посилання в програмі.

### Літерали list і dict

Літерали list і dict компілюються в `OpMakeList` та `OpMakeDict` з попереднім push елементів.

### Літерали структур

Літерали структур компілюються в `OpMakeStruct` з push полів у оголошеному порядку.

Для узагальнених літералів структур компілятор резолвить мономорфізоване ім'я структури та генерує `OpMakeStruct` для індексу конкретного типу.

### Виклики методів

Виклики методів компілюються з коректною обробкою receiver-а:

- Методи екземпляра: вираз receiver-а компілюється першим і передається як перший аргумент
- Статичні методи: компілюються як звичайні виклики функцій без receiver-а
- Компілятор виявляє методи екземпляра через поле `Method.Decl` у binding-ах
- Параметри за замовчуванням обробляються через переупорядкування аргументів у `reorderCallArgs`

### Індексація типів структур

Компілятор будує карту `structIndex` від імен структур до індексів типів:
- Дублікати імен структур між модулями пропускаються для запобігання паніці index-out-of-range
- Реєструється лише перше входження кожного імені структури
- Це забезпечує консистентну індексацію при наявності однойменних структур у різних модулях

### Інтерпольовані рядки

`"x=${expr}"` lowering:

1. Компіляція кожної частини
2. `OpStringify` для частин-виразів
3. `OpConcatString` між частинами

### Try/Catch

`try { ... } catch (...) { ... }` компілюється в:

1. `OpBeginTry` з IP обробника
2. Інструкції try-блоку
3. `OpEndTry`
4. Блок обробника

При наявності кількох типізованих catch-клауз блок обробника генерує ланцюжок:

1. Для кожного `catch (var | StructType)`:
   - `OpIsStructType A` (A = індекс типу структури) — peek TOS, push bool
   - `OpJumpIfFalse` до наступної клаузи
   - Збереження thrown-значення в локальну змінну, тіло клаузи, `OpJump` в кінець
2. Для `catch (var | error)`: безумовна fallback-клауза
3. Фінальний `OpThrow` перекидає необроблені помилки

Кожна клауза має власну локальну область видимості для запобігання перевизначенню змінних.

### Switch / Continue

- `switch` lowering: перевірки рівності (`OpEq`) + умовні переходи (`OpJumpIfFalse`) для кожного `case`.
- Тіло кожного matched case закінчується `OpJump` для пропуску решти клауз.
- `continue` lowering: перехід до loop-специфічної цілі continue.

### Опціональне ланцюжування

Опціональні ланцюжки (`?.`) використовують `OpJumpIfNone`:

1. Обчислення receiver-а/callee.
2. `OpJumpIfNone` переходить до none-шляху, коли значення — `none`.
3. Non-none шлях обчислює член/виклик та обгортає результат через `OpMakeSome`.

### Defer

`defer` lowering в `OpPushDefer`:

1. Обчислення та захоплення аргументів відкладеного виклику.
2. Обчислення відкладеного callee.
3. Генерація `OpPushDefer` із кількістю захоплених аргументів.

Відкладені виклики виконуються пізніше при обробці return у VM (порядок LIFO).

### Async/Await Lowering

Async-метадані та опкоди генеруються наступним чином:

- `ast.FunDecl.IsAsync` проставляється в `ir.Function.IsAsync`.
- Прямий виклик оголошення async-функції lowering в `OpSpawn`.
- `await expr` lowering в `OpAwait` після компіляції `expr`.

`OpSpawn` і `OpAwait` — async-границі на рівні VM:

1. `OpSpawn` споживає скомпільовані аргументи виклику і кладе в стек `Future`.
2. `OpAwait` споживає `Future`:
   - ready + success: кладе розв'язане значення в стек
   - ready + failure: викидає/прокидає помилку
   - not ready: призупиняє поточний async-контекст задачі

У поточній реалізації `OpSpawn` виконує цільове замикання одразу та обгортає завершення/помилку в `Future`; поведінка suspend/resume забезпечується через `OpAwait` + координацію scheduler/event-loop.

## Приклад

Вихідний код:

```avenir
print("a" + "b");
```

Lowered IR (концептуально):

```
OpConst "a"
OpConst "b"
OpConcatString
OpCallBuiltin print, 1
```

## Посилання

- `internal/ir/ir.go`
- `internal/ir/compiler.go`
