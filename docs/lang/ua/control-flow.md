# Керування потоком

Avenir підтримує умовні конструкції, цикли та обробку винятків.

## Умови

### `if`

```avenir
if (condition) {
}
```

Умова повинна мати тип `bool`.

### `if` / `else`

```avenir
if (condition) {
} else {
}
```

### `else if`

Можна ланцюжком перевіряти кілька умов.

### Кілька умов

У `if` можна використовувати `;` як скорочення для `&&`:

```avenir
if (x > 0; x < 10) {
}
```

## Цикли

### `while`

```avenir
while (condition) {
}
```

### `for`

Підтримується цикл у стилі C:

```avenir
for (init; condition; post) {
}
```

Усі частини можуть бути пропущені.

### `for (item in list)`

Цикл перебору ітерується списком:

```avenir
for (item in list) {
    print(item);
}
```

### `break`

`break` завершує найближчий цикл і може використовуватись лише всередині циклів.

## Винятки

### `try` / `catch`

```avenir
try {
} catch (e | error) {
}
```

Змінна `catch` може мати тип `error` або тип структури.

### Типізовані catch-клаузи

`try` підтримує кілька типізованих catch-клауз для зіставлення з конкретними
типами помилок. Клаузи перевіряються по порядку:

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
        print("не знайдено: " + e.path);
    } catch (e | PermissionDenied) {
        print("відмовлено: " + e.file);
    } catch (e | error) {
        print("інша помилка");
    }
}
```

`catch (e | error)` — fallback-клауза, що перехоплює будь-яку помилку.

### `throw`

```avenir
throw error("щось пішло не так");
```

Вираз може бути типу `error` або оголошеного типу структури.

### Декларація throws

Функції можуть оголошувати типи помилок, що викидаються, через `!` після типу
повернення:

```avenir
fun readFile(path | string) | string ! FileNotFound {
    throw FileNotFound{path = path};
}
```

Type checker валідує, що вирази `throw` всередині тіла функції відповідають
оголошеним типам.

## Блоки

Блоки `{ ... }` створюють нову область видимості.

## Вирази як оператори

Вираз можна використовувати як окремий оператор, його результат буде відкинуто.
