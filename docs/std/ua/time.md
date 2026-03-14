# std.time — Час і тривалість

Модуль `std.time` надає утиліти для роботи з датою/часом та тривалістю.

## Імпорт

```avenir
import std.time;
```

## Структури

### Duration

```avenir
struct Duration {
    _nanos | int
}
```

### DateTime

```avenir
struct DateTime {
    _handle | any
}
```

## Функції

### Поточний час

| Функція | Опис |
|---------|------|
| `now() \| DateTime` | Поточна дата/час |
| `nowUnix() \| int` | Unix-мітка часу (секунди) |
| `nowUnixMilli() \| int` | Unix-мітка часу (мілісекунди) |
| `nowUnixNano() \| int` | Unix-мітка часу (наносекунди) |

### Сон

| Функція | Опис |
|---------|------|
| `sleep(nanos \| int) \| void` | Синхронний сон (наносекунди) |
| `asyncSleep(nanos \| int) \| Future<void>` | Асинхронний сон |

### Парсинг і форматування

| Функція | Опис |
|---------|------|
| `parse(value \| string, format \| string) \| DateTime` | Парсинг рядка в DateTime |
| `formatUnix(timestamp \| int, format \| string) \| string` | Форматування Unix-мітки |

### Конструктори Duration

| Функція | Опис |
|---------|------|
| `fromNanos(n \| int) \| Duration` | Тривалість з наносекунд |
| `fromMicros(n \| int) \| Duration` | Тривалість з мікросекунд |
| `fromMillis(n \| int) \| Duration` | Тривалість з мілісекунд |
| `fromSeconds(n \| int) \| Duration` | Тривалість із секунд |
| `fromMinutes(n \| int) \| Duration` | Тривалість з хвилин |
| `fromHours(n \| int) \| Duration` | Тривалість з годин |

### Таймаут

| Функція | Опис |
|---------|------|
| `withTimeout(f \| Future<any>, d \| Duration) \| Future<any>` | Обмежити час очікування future |

## Методи DateTime

| Метод | Опис |
|-------|------|
| `year() \| int` | Рік |
| `month() \| int` | Місяць (1-12) |
| `day() \| int` | День місяця |
| `hour() \| int` | Година (0-23) |
| `minute() \| int` | Хвилина (0-59) |
| `second() \| int` | Секунда (0-59) |
| `weekday() \| int` | День тижня (0=неділя) |
| `unix() \| int` | Unix-мітка часу |
| `add(d \| Duration) \| DateTime` | Додати тривалість |
| `sub(d \| Duration) \| DateTime` | Відняти тривалість |
| `format(fmt \| string) \| string` | Форматувати в рядок |

## Методи Duration

| Метод | Опис |
|-------|------|
| `hours() \| float` | Отримати години |
| `minutes() \| float` | Отримати хвилини |
| `seconds() \| float` | Отримати секунди |
| `millis() \| int` | Отримати мілісекунди |
| `nanos() \| int` | Отримати наносекунди |
| `add(d \| Duration) \| Duration` | Додати тривалість |
| `sub(d \| Duration) \| Duration` | Відняти тривалість |

## Токени формату

| Токен | Опис | Приклад |
|-------|------|---------|
| `YYYY` | 4-значний рік | `2024` |
| `MM` | Місяць (з нулем) | `01`–`12` |
| `DD` | День (з нулем) | `01`–`31` |
| `hh` | Година (з нулем) | `00`–`23` |
| `mm` | Хвилина (з нулем) | `00`–`59` |
| `ss` | Секунда (з нулем) | `00`–`59` |

## Приклади

### Базове використання

```avenir
import std.time;

fun main() | void {
    var now | std.time.DateTime = std.time.now();
    print(now.format("YYYY-MM-DD hh:mm:ss"));
    print(now.year());
    print(now.month());
}
```

### Арифметика з тривалістю

```avenir
import std.time;

fun main() | void {
    var d | std.time.Duration = std.time.fromHours(2);
    var d2 | std.time.Duration = d.add(std.time.fromMinutes(30));
    print(d2.hours());
}
```

### Async sleep з таймаутом

```avenir
import std.time;

async fun main() | void {
    await std.time.asyncSleep(std.time.fromSeconds(1).nanos());
    print("Прокинулися!");
}
```

## Блокуюча vs неблокуюча поведінка

- `sleep()` блокує поточний потік
- `asyncSleep()` виконується через async event loop і повертає future

## Обробка помилок

Функції парсингу та форматування можуть викидати помилки при невалідному введенні:

```avenir
try {
    var dt | std.time.DateTime = std.time.parse("invalid", "YYYY-MM-DD");
} catch (e | error) {
    print("Помилка парсингу: " + e.message());
}
```
