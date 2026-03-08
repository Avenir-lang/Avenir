# Runtime и Builtins

Этот документ описывает runtime-сервисы и подсистему встроенных функций.

## Обзор

Runtime связывает исполнение VM с сервисами хоста (I/O, сеть, файловая система,
HTTP). Builtins регистрируются в Go и экспортируются в Avenir как функции/методы.

Ключевые файлы:

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

## Окружение (`Env`)

`runtime.Env` предоставляет:

- IO (`Println`, `ReadLine`)
- сеть (`Connect`, `Listen`, `Accept`, `Read`, `Write`, `Close`)
- FS (`Open`, `Read`, `Write`, `Close`, `Exists`, `Remove`, `Mkdir`)
- HTTP (`Request`, `Listen`, `Accept`, `Respond`)
- `ExecRoot()` для разрешения путей

## Реестр builtins

Builtins регистрируются через `builtins.Register` в `init()`.
Runtime импортирует пакеты builtins для автрорегистрации.

## Асинхронные компоненты runtime

Async-backend использует отдельные примитивы:

- `Future` (`internal/runtime/future.go`)
  - поля: `Ready`, `Result`, `Err`, список waiters
  - `Resolve`/`Reject` завершают future и будят ожидающие задачи
  - регистрация waiters защищена `sync.Mutex`
- `Task` (`internal/runtime/task.go`)
  - поля: `ID`, `Status`, `Future`, `Scheduler`, `StepFn`
  - статусы: `TaskReady`, `TaskRunning`, `TaskSuspended`, `TaskDone`, `TaskFailed`
- `Scheduler` (`internal/runtime/scheduler.go`)
  - поддерживает очередь ready-задач и map suspended-задач
  - раздаёт ID задачам и пере-планирует waiters
- `RunEventLoop` (`internal/runtime/eventloop.go`)
  - выполняет задачи из ready-очереди
  - переводит failed-задачи в rejected future
  - сообщает о deadlock, если остались только suspended-задачи

### Поток ожидания (waiter flow)

Когда VM выполняет `OpAwait` для неготового future (в async-контексте):

1. VM регистрирует задачу через `Future.AddWaiter(task)`.
2. Scheduler переводит задачу в suspended.
3. При `Resolve`/`Reject` future возвращает waiter-задачи в ready-очередь через `Scheduler.Schedule`.
4. Event loop снова выбирает их на выполнение.

## Интеграция стандартной библиотеки

Stdlib Avenir (`./std`) написана на Avenir, но обращается к внутренним builtins:

- `std.fs` → `__builtin_fs_*`
- `std.net` → `__builtin_socket_*`
- `std.json` → `__builtin_json_*`
- `std.http` → `__builtin_http_*`
- `std.time` → `__builtin_time_*`

## Ссылки

- `internal/runtime/runtime.go`
- `internal/runtime/env.go`
- `internal/runtime/builtins/registry.go`

