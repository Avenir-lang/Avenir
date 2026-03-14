# Runtime та Builtins

Цей документ описує runtime-сервіси та підсистему вбудованих функцій.

## Огляд

Runtime з'єднує виконання VM з сервісами хоста (I/O, мережа, файлова система, HTTP). Builtins реєструються в Go і доступні в Avenir як функції або методи.

Ключові файли:

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

## Оточення (`Env`)

`runtime.Env` надає:

- IO (`Println`, `ReadLine`)
- Мережу (`Connect`, `Listen`, `Accept`, `Read`, `Write`, `Close`)
- FS (`Open`, `Read`, `Write`, `Close`, `Exists`, `Remove`, `Mkdir`)
- HTTP (`Request`, `Listen`, `Accept`, `Respond`)
- `ExecRoot()` для резолву шляхів
- Хуки виклику замикань для функцій списків (`map`, `filter`, `reduce`)

VM конфігурує оточення при запуску, включно з іменами типів структур і caller-ом замикань.

## Реєстр builtins

Builtins реєструються через `builtins.Register` у функціях `init()`. Runtime імпортує всі пакети builtins для автоматичної реєстрації:

```go
_ "avenir/internal/runtime/builtins/bytes"
_ "avenir/internal/runtime/builtins/collections"
// ...
```

Кожен builtin надає:

- `Meta` (ім'я, арність, типи параметрів, тип receiver-а)
- `Call` (реалізація)

### Функції vs Методи

- Функції використовують `ReceiverType = TypeVoid`.
- Методи використовують `ReceiverType` та `MethodName` для пошуку за типом receiver-а.

## Потік виклику

VM → `runtime.CallBuiltin`:

1. Пошук builtin за ID
2. Конвертація `[]value.Value` → `[]interface{}`
3. Виклик builtin з `Env`
4. Конвертація результату назад у `value.Value`

Помилки від builtins прокидаються через VM і перехоплюються `try / catch`.

## Інтеграція стандартної бібліотеки

Stdlib Avenir (`./std`) написана на Avenir, але делегує внутрішнім builtins для доступу до хоста:

- `std.fs` → `__builtin_fs_*`
- `std.net` → `__builtin_socket_*`
- `std.json` → `__builtin_json_*`
- `std.http` → `__builtin_http_*`
- `std.time` → `__builtin_time_*`

Ці builtins не є частиною публічного API мови, але є стабільними внутрішніми інтерфейсами, що використовуються std-модулями.

## Асинхронні компоненти runtime

Async-виконання забезпечується виділеними runtime-примітивами:

- `Future` (`internal/runtime/future.go`)
  - поля: `Ready`, `Result`, `Err`, список waiters
  - `Resolve`/`Reject` позначають завершення і будять задачі-очікувачі
  - реєстрація waiters захищена `sync.Mutex`
- `Task` (`internal/runtime/task.go`)
  - поля: `ID`, `Status`, `Future`, `Scheduler`, `StepFn`
  - стани: `TaskReady`, `TaskRunning`, `TaskSuspended`, `TaskDone`, `TaskFailed`
- `Scheduler` (`internal/runtime/scheduler.go`)
  - підтримує ready-чергу + map suspended-задач
  - видає ID задачам і переплановує waiters
- `RunEventLoop` (`internal/runtime/eventloop.go`)
  - повторно виконує ready-задачі
  - переводить failed-задачі у rejected future
  - повідомляє про deadlock, якщо залишилися лише suspended-задачі

### Потік очікування (waiter flow)

Коли VM виконує `OpAwait` для неготового future (в async-контексті задачі):

1. VM викликає `Future.AddWaiter(task)`.
2. Задача призупиняється scheduler-ом.
3. При `Resolve`/`Reject` future переплановує waiter-задачі через `Scheduler.Schedule`.
4. Event loop підхоплює відновлені задачі з ready-черги.

### Потокобезпека

`Scheduler` захищений `sync.Mutex`, оскільки горутини (від завершення async I/O) викликають `Schedule` конкурентно з читанням ready-черги event loop-ом. Event loop використовує `IsIdle()` для атомарної перевірки порожнечі, уникаючи TOCTOU-гонок між `HasTasks`/`HasSuspended`.

### Конкурентність OpSpawn

Коли компілятор зустрічає виклик `async fun`, він генерує `OpSpawn` замість `OpCall`. VM обробляє `OpSpawn` так:

1. Pop аргументів з батьківського стеку
2. Створення нового `Future` для результату spawn
3. Створення **дочірньої VM** через `spawnChild()` — ділить `mod`, `env` та `scheduler`, але має власний стек і фрейми
4. Push аргументів у стек дочірньої VM
5. Створення `Task`, чий `StepFn` викликає `childVM.callClosure`
6. Планування дочірньої задачі на спільному scheduler-і
7. Push `Future` в батьківський стек

Дочірня задача працює кооперативно з батьківською та іншими задачами через спільний event loop. Коли дочірня завершується, вона розв'язує свій future, будячи будь-яку задачу, що його очікувала.

### Асинхронні builtins

Async builtins реєструються з полем `CallAsync`, що повертає `*AsyncHandle`. VM використовує `OpCallBuiltinAsync` для їх виклику:

1. Виклик `runtime.CallBuiltinAsync(env, id, args)` → повертає `*AsyncHandle`
2. Створення `Future` та з'єднання через `ah.WireToFuture(fut)`
3. Push future в стек

`AsyncHandle` виконує фактичний I/O в Go-горутині. Після завершення з'єднаний future розв'язується/відхиляється, плануючи будь-які задачі, що очікують.

Категорії async builtins:
- **FS**: `__builtin_async_fs_open`, `_read`, `_read_all`, `_write`, `_close`, `_exists`, `_remove`, `_mkdir`
- **Net**: `__builtin_async_socket_connect`, `_read`, `_write`, `_close`, `_accept`
- **HTTP**: `__builtin_async_http_request`, `_accept`, `_respond`
- **Time**: `__builtin_async_time_sleep`

## Exec Root і резолв шляхів

Runtime-оточення надає `ExecRoot()` для резолву відносних шляхів у `std.fs`. Він встановлюється на директорію вхідного `.av` файлу, щоб шляхи користувача резолвилися відносно виконуваної програми, а не робочої директорії компілятора.

## Примітки

- Builtins оперують на `value.Value` і повертають `value.Value`.
- Builtins не повинні мутувати значення list; методи dict можуть мутувати на місці.

## Посилання

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

