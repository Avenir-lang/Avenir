# Runtime та Builtins

Цей документ описує runtime-сервіси та підсистему вбудованих функцій.

## Огляд

Runtime з'єднує виконання VM з сервісами хоста (I/O, мережа, файлова система,
HTTP). Builtins реєструються в Go і доступні в Avenir як функції/методи.

Ключові файли:

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

## Оточення (`Env`)

`runtime.Env` надає:

- IO (`Println`, `ReadLine`)
- мережу (`Connect`, `Listen`, `Accept`, `Read`, `Write`, `Close`)
- FS (`Open`, `Read`, `Write`, `Close`, `Exists`, `Remove`, `Mkdir`)
- HTTP (`Request`, `Listen`, `Accept`, `Respond`)
- `ExecRoot()` для резолву шляхів

## Реєстр builtins

Builtins реєструються через `builtins.Register` у `init()`.
Runtime імпортує пакети builtins для автоматичної реєстрації.

## Асинхронні компоненти runtime

Async-backend використовує окремі примітиви:

- `Future` (`internal/runtime/future.go`)
  - поля: `Ready`, `Result`, `Err`, список waiters
  - `Resolve`/`Reject` завершують future і будять задачі-очікувачі
  - реєстрація waiters захищена `sync.Mutex`
- `Task` (`internal/runtime/task.go`)
  - поля: `ID`, `Status`, `Future`, `Scheduler`, `StepFn`
  - стани: `TaskReady`, `TaskRunning`, `TaskSuspended`, `TaskDone`, `TaskFailed`
- `Scheduler` (`internal/runtime/scheduler.go`)
  - підтримує ready-чергу і map suspended-задач
  - видає ID задачам і переплановує waiters
- `RunEventLoop` (`internal/runtime/eventloop.go`)
  - виконує задачі з ready-черги
  - переводить failed-задачі у rejected future
  - повідомляє про deadlock, якщо лишилися тільки suspended-задачі

### Потік очікування (waiter flow)

Коли VM виконує `OpAwait` для неготового future (в async-контексті):

1. VM реєструє задачу через `Future.AddWaiter(task)`.
2. Scheduler переводить задачу в suspended.
3. Під час `Resolve`/`Reject` future повертає waiter-задачі в ready-чергу через `Scheduler.Schedule`.
4. Event loop знову вибирає їх на виконання.

## Інтеграція стандартної бібліотеки

Stdlib Avenir (`./std`) написана на Avenir, але звертається до внутрішніх builtins:

- `std.fs` → `__builtin_fs_*`
- `std.net` → `__builtin_socket_*`
- `std.json` → `__builtin_json_*`
- `std.http` → `__builtin_http_*`
- `std.time` → `__builtin_time_*`

## Посилання

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

