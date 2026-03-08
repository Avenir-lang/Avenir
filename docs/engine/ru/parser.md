# Парсер

Этот документ описывает рекурсивный descent-парсер Avenir,
реализованный в `internal/parser/parser.go`.

## Обзор

Парсинг выполняется за один проход по потоку токенов с lookahead.
Парсер строит узлы AST из `internal/ast` и накапливает ошибки,
а не завершает работу panic-ом.

## Структура программы

На верхнем уровне парсер ожидает:

1. декларацию `pckg`
2. декларации `import`
3. декларации `struct`
4. декларации `interface`
5. декларации `fun` (включая методы)

## Приоритет выражений

Выражения разбираются каскадом функций по приоритетам:

1. `parseOr` → `||`
2. `parseAnd` → `&&`
3. `parseEquality` → `==`, `!=`
4. `parseRelational` → `<`, `<=`, `>`, `>=`
5. `parseAdditive` → `+`, `-`
6. `parseMultiplicative` → `*`, `/`, `%`
7. `parseUnary` → `!`, унарный `-`, `await`
8. `parsePostfix` → доступ к полю, вызовы, индексация
9. `parsePrimary` → литералы, идентификаторы, группировка

## Операторы

Поддерживаются:

- объявление переменных: `var name | Type = expr;`
- присваивания: `name = expr;`
- expression statements
- `if` / `else`, `while`, `for`, `for (item in list)`
- `return`, `try` / `catch`, `throw`, `break`, `continue`
- `switch` / `case` / `default`
- `defer` (только call expression)

## Парсинг асинхронного синтаксиса

### Async-функции

`parseFunDecl` принимает модификаторы в порядке:

1. `pub` (опционально)
2. `async` (опционально)
3. `fun` (обязательно)

Флаг асинхронности сохраняется в `ast.FunDecl.IsAsync`.

### Await-выражения

`await` разбирается в `parseUnary`, то есть имеет приоритет унарного оператора.

Пример:

```avenir
await a + b
```

разбирается как:

```avenir
(await a) + b
```

Парсер строит `ast.AwaitExpr`, а проверка типов происходит в type checker-е.

## Вызовы и доступ к членам

`parsePostfix` обрабатывает:

- `expr.name`
- `expr?.name`
- `expr(args...)`
- `expr?.(args...)`, `expr?.method(args...)`
- `expr[index]`

Именованные аргументы (`name = expr`) распознаются только в списках аргументов вызова.

## Литералы и типы

Парсер поддерживает литералы `int`, `float`, `string`, `bytes`, `bool`,
`none`, `some(expr)`, list/dict/struct литералы и интерполированные строки.

Типы: простые, qualified, list, dict, optional (`T?`), union, function.

## Обработка ошибок

Ошибки накапливаются через `p.errorf(...)`, `p.Errors()` возвращает полный список.

## Ссылки

- `internal/parser/parser.go`
- `internal/ast/ast.go`
