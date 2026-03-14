# std.fs — Файлова система

Модуль `std.fs` надає доступ до файлової системи.

## Імпорт

```avenir
import std.fs;
```

## Структура File

```avenir
struct File {
    _handle | any
}
```

Файли створюються через `open` або `asyncOpen` — прямий доступ до `_handle` не рекомендується.

## Функції

### Синхронні

| Функція | Опис |
|---------|------|
| `open(path \| string, mode \| string) \| File` | Відкрити файл з режимом |
| `exists(path \| string) \| bool` | Перевірити існування файлу |
| `remove(path \| string) \| void` | Видалити файл |
| `mkdir(path \| string) \| void` | Створити директорію |

### Асинхронні

| Функція | Опис |
|---------|------|
| `asyncOpen(path \| string, mode \| string) \| Future<File>` | Асинхронне відкриття файлу |
| `asyncExists(path \| string) \| Future<bool>` | Асинхронна перевірка існування |
| `asyncRemove(path \| string) \| Future<void>` | Асинхронне видалення файлу |
| `asyncMkdir(path \| string) \| Future<void>` | Асинхронне створення директорії |

### Хелпери шляхів

| Функція | Опис |
|---------|------|
| `joinPath(parts \| list<string>) \| string` | Об'єднання частин шляху |
| `baseName(path \| string) \| string` | Ім'я файлу з шляху |
| `dirName(path \| string) \| string` | Директорія з шляху |

## Методи File

### Синхронні

| Метод | Опис |
|-------|------|
| `read(size \| int) \| bytes` | Прочитати до `size` байтів |
| `readAll() \| bytes` | Прочитати весь вміст |
| `readString() \| string` | Прочитати весь вміст як рядок |
| `write(data \| bytes) \| int` | Записати байти, повертає кількість записаних |
| `writeString(data \| string) \| int` | Записати рядок |
| `close() \| void` | Закрити файл |

### Асинхронні

| Метод | Опис |
|-------|------|
| `asyncRead(size \| int) \| Future<bytes>` | Асинхронне читання |
| `asyncReadAll() \| Future<bytes>` | Асинхронне читання всього вмісту |
| `asyncReadString() \| Future<string>` | Асинхронне читання як рядок |
| `asyncWrite(data \| bytes) \| Future<int>` | Асинхронний запис |
| `asyncWriteString(data \| string) \| Future<int>` | Асинхронний запис рядка |
| `asyncClose() \| Future<void>` | Асинхронне закриття |

## Режими відкриття

| Режим | Опис |
|-------|------|
| `"r"` | Читання (за замовчуванням) |
| `"w"` | Запис (створює/очищує) |
| `"a"` | Дозапис |
| `"rw"` | Читання та запис |

## Приклад

```avenir
import std.fs;

fun main() | void {
    var f | std.fs.File = std.fs.open("hello.txt", "w");
    f.writeString("Hello, World!");
    f.close();

    var f2 | std.fs.File = std.fs.open("hello.txt", "r");
    var content | string = f2.readString();
    f2.close();
    print(content);
}
```

### Асинхронний приклад

```avenir
import std.fs;

async fun main() | void {
    var f | std.fs.File = await std.fs.asyncOpen("data.txt", "r");
    var content | string = await f.asyncReadString();
    await f.asyncClose();
    print(content);
}
```

## Резолв шляхів

Відносні шляхи резолвляться від директорії вхідного `.av` файлу (через `ExecRoot()`), а не від робочої директорії компілятора.

## Обробка помилок

Файлові операції можуть викидати помилки (файл не знайдено, доступ заборонено тощо). Використовуйте `try/catch` для обробки:

```avenir
try {
    var f | std.fs.File = std.fs.open("missing.txt", "r");
} catch (e | error) {
    print("Помилка файлу: " + e.message());
}
```

## Блокуюча vs неблокуюча поведінка

- Синхронні функції (`open`, `read`, `write`) блокують поточний потік
- Асинхронні функції (`asyncOpen`, `asyncRead`, `asyncWrite`) виконують I/O у фонових горутинах і повертають future
