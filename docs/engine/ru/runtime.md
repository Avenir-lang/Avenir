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

- `Future` (`internal/runtime/future.go`): хранит `Ready`, `Result`, `Err`, список waiters
- `Task` (`internal/runtime/task.go`): статус задачи и шаг `func() (TaskStatus, error)`
- `Scheduler` (`internal/runtime/scheduler.go`): очередь готовых задач и множество suspended
- `EventLoop` (`internal/runtime/eventloop.go`): цикл исполнения, deadlock detection, propagation ошибок

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

