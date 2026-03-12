# Управление потоком

Avenir поддерживает условные конструкции, циклы и обработку исключений.

## Условия

### `if`

```avenir
if (condition) {
}
```

Условие должно иметь тип `bool`.

### `if` / `else`

```avenir
if (condition) {
} else {
}
```

### `else if`

Можно цепочкой проверять несколько условий.

### Несколько условий

В `if` можно использовать `;` как сокращение для `&&`:

```avenir
if (x > 0; x < 10) {
}
```

## Циклы

### `while`

```avenir
while (condition) {
}
```

### `for`

Поддерживается цикл в стиле C:

```avenir
for (init; condition; post) {
}
```

Все части могут быть опущены.

### `for (item in list)`

Цикл перебора итерирует по списку:

```avenir
for (item in list) {
    print(item);
}
```

### `break`

`break` завершает ближайший цикл и может использоваться только внутри циклов.

## Исключения

### `try` / `catch`

```avenir
try {
} catch (e | error) {
}
```

Переменная `catch` может иметь тип `error` или тип структуры.

### Типизированные catch-клаузы

`try` поддерживает несколько типизированных catch-клауз для сопоставления с
конкретными типами ошибок. Клаузы проверяются по порядку:

```avenir
struct FileNotFound {
    path | string;
}

struct PermissionDenied {
    file | string;
}

fun riskyOp() | void ! FileNotFound, PermissionDenied {
    throw FileNotFound{path = "/tmp/missing.txt"};
}

fun main() | void {
    try {
        riskyOp();
    } catch (e | FileNotFound) {
        print("не найден: " + e.path);
    } catch (e | PermissionDenied) {
        print("отказано: " + e.file);
    } catch (e | error) {
        print("другая ошибка");
    }
}
```

`catch (e | error)` — fallback-клауза, перехватывающая любую ошибку.

### `throw`

```avenir
throw error("что-то пошло не так");
```

Выражение может быть типа `error` или объявленного типа структуры.

### Декларация throws

Функции могут объявлять типы выбрасываемых ошибок через `!` после типа возврата:

```avenir
fun readFile(path | string) | string ! FileNotFound {
    throw FileNotFound{path = path};
}
```

Type checker валидирует, что выражения `throw` внутри тела функции соответствуют
объявленным типам.

## Блоки

Блоки `{ ... }` создают новую область видимости.

## Выражения как операторы

Выражение можно использовать как отдельный оператор, его результат будет отброшен.
