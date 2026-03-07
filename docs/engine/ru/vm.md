# Виртуальная машина

Этот документ описывает VM Avenir в `internal/vm/vm.go`.

## Обзор

VM — стековый интерпретатор IR-байткода. Он исполняет функции,
скомпилированные в `internal/ir`, и использует runtime-окружение
для builtins и хостовых сервисов.

## Представление значений

Значения хранятся как `value.Value` (см. `internal/value`) с дискриминатором `Kind`:

- `Int`, `Float`, `String`, `Bool`, `Bytes`
- `List`, `Dict`, `Struct`
- `Optional` (`some` / `none`)
- `Closure`, `Future`, `Error`

`Value.String()` используется для строкового представления и печати.

## Стек и фреймы

VM поддерживает:

- стек значений (`stack` + `sp`)
- стек вызовов (`[]Frame`)

Каждый `Frame` хранит текущую функцию/замыкание, `IP` и `Base`.

## Цикл исполнения

VM читает инструкции текущего фрейма и выполняет их в цикле.
Ключевые инструкции:

- `OpConst`, `OpLoadLocal`, `OpStoreLocal`
- арифметика и сравнения
- `OpCall`, `OpCallValue`, `OpCallBuiltin`
- `OpSpawn`, `OpAwait` (асинхронный backend)

## Асинхронное исполнение

Async-модель кооперативная и однопоточная.

- `RunMain` проверяет `Function.IsAsync`.
- Для async `main` создаются `Scheduler` и `EventLoop`.
- Вызов async-функции даёт `Future`, `await` ожидает результат.

Если `await` встречает незавершённый `Future`, текущая задача приостанавливается,
её состояние (стек/фреймы/handlers) сохраняется, и планировщик переключается
на другие готовые задачи.

## Ошибки и исключения

Runtime-ошибки преобразуются в `error`-значения:

- `raiseError(err)` оборачивает ошибку в `value.ErrorValue`
- `throwValue` разматывает стек до ближайшего обработчика (`OpBeginTry`)

## Ссылки

- `internal/vm/vm.go`
- `internal/value/value.go`

