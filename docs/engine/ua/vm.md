# Віртуальна машина

Цей документ описує VM Avenir у `internal/vm/vm.go`.

## Огляд

VM — стековий інтерпретатор IR-байткоду. Вона виконує функції,
скомпільовані в `internal/ir`, і використовує runtime-оточення
для builtins та хостових сервісів.

## Представлення значень

Значення зберігаються як `value.Value` (див. `internal/value`) з `Kind`:

- `Int`, `Float`, `String`, `Bool`, `Bytes`
- `List`, `Dict`, `Struct`
- `Optional` (`some` / `none`)
- `Closure`, `Future`, `Error`

`Value.String()` використовується для рядкового представлення і друку.

## Стек і фрейми

VM підтримує:

- стек значень (`stack` + `sp`)
- стек викликів (`[]Frame`)

Кожен `Frame` містить поточну функцію/замикання, `IP` і `Base`.

## Цикл виконання

VM читає інструкції поточного фрейма і виконує їх у циклі.
Ключові інструкції:

- `OpConst`, `OpLoadLocal`, `OpStoreLocal`
- арифметика та порівняння
- `OpCall`, `OpCallValue`, `OpCallBuiltin`
- `OpSpawn`, `OpAwait` (асинхронний backend)

## Асинхронне виконання

Async-модель кооперативна і однопотокова.

- `RunMain` перевіряє `Function.IsAsync`.
- Для async `main` створюються `Scheduler` і `EventLoop`.
- Виклик async-функції повертає `Future`, `await` очікує результат.

Якщо `await` отримує незавершений `Future`, поточна задача призупиняється,
її стан (стек/фрейми/handlers) зберігається, а планувальник перемикається
на інші готові задачі.

## Помилки та виключення

Runtime-помилки перетворюються на `error`-значення:

- `raiseError(err)` обгортає помилку у `value.ErrorValue`
- `throwValue` розмотує стек до найближчого обробника (`OpBeginTry`)

## Посилання

- `internal/vm/vm.go`
- `internal/value/value.go`

