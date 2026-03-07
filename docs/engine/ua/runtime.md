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

- `Future` (`internal/runtime/future.go`): зберігає `Ready`, `Result`, `Err`, список waiters
- `Task` (`internal/runtime/task.go`): статус задачі і крок `func() (TaskStatus, error)`
- `Scheduler` (`internal/runtime/scheduler.go`): черга готових задач і множина suspended
- `EventLoop` (`internal/runtime/eventloop.go`): цикл виконання, deadlock detection, propagation помилок

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

