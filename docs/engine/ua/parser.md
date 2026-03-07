# Парсер

Цей документ описує рекурсивний descent-парсер Avenir,
реалізований у `internal/parser/parser.go`.

## Огляд

Парсинг виконується за один прохід по потоку токенів з lookahead.
Парсер будує вузли AST з `internal/ast` і накопичує помилки,
а не завершує роботу panic-ом.

## Структура програми

На верхньому рівні парсер очікує:

1. декларацію `pckg`
2. декларації `import`
3. декларації `struct`
4. декларації `interface`
5. декларації `fun` (включно з методами)

## Пріоритет виразів

Вирази розбираються каскадом функцій за пріоритетами:

1. `parseOr` → `||`
2. `parseAnd` → `&&`
3. `parseEquality` → `==`, `!=`
4. `parseRelational` → `<`, `<=`, `>`, `>=`
5. `parseAdditive` → `+`, `-`
6. `parseMultiplicative` → `*`, `/`, `%`
7. `parseUnary` → `!`, унарний `-`
8. `parsePostfix` → доступ до поля, виклики, індексація
9. `parsePrimary` → літерали, ідентифікатори, групування

## Оператори

Підтримуються:

- оголошення змінних: `var name | Type = expr;`
- присвоєння: `name = expr;`
- expression statements
- `if` / `else`, `while`, `for`, `for (item in list)`
- `return`, `try` / `catch`, `throw`, `break`

## Виклики і доступ до членів

`parsePostfix` обробляє:

- `expr.name`
- `expr(args...)`
- `expr[index]`

Іменовані аргументи (`name = expr`) розпізнаються лише у списках аргументів виклику.

## Літерали і типи

Парсер підтримує літерали `int`, `float`, `string`, `bytes`, `bool`,
`none`, `some(expr)`, list/dict/struct літерали та інтерпольовані рядки.

Типи: прості, qualified, list, dict, optional (`T?`), union, function.

## Обробка помилок

Помилки накопичуються через `p.errorf(...)`, `p.Errors()` повертає повний список.

## Посилання

- `internal/parser/parser.go`
- `internal/ast/ast.go`
