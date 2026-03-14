# std.http — HTTP клієнт і сервер

Модуль `std.http` надає мінімальний API для HTTP клієнта та сервера.

## Імпорт

```avenir
import std.http.client;
import std.http.server;
```

## Структури

### HttpRequest (клієнт)

```avenir
struct HttpRequest {
    method  | string
    url     | string
    headers | dict<string>
    body    | string
}
```

### HttpResponse (клієнт)

```avenir
struct HttpResponse {
    status  | int
    headers | dict<string>
    body    | string
}
```

## Функції клієнта

### Синхронні

| Функція | Опис |
|---------|------|
| `get(url \| string) \| string` | GET-запит, повертає тіло відповіді |
| `post(url \| string, body \| string) \| string` | POST-запит |
| `put(url \| string, body \| string) \| string` | PUT-запит |
| `delete(url \| string) \| string` | DELETE-запит |
| `request(req \| HttpRequest) \| HttpResponse` | Повний HTTP-запит |

### Асинхронні

| Функція | Опис |
|---------|------|
| `asyncGet(url \| string) \| Future<string>` | Асинхронний GET |
| `asyncPost(url \| string, body \| string) \| Future<string>` | Асинхронний POST |
| `asyncPut(url \| string, body \| string) \| Future<string>` | Асинхронний PUT |
| `asyncDelete(url \| string) \| Future<string>` | Асинхронний DELETE |
| `asyncRequest(req \| HttpRequest) \| Future<HttpResponse>` | Асинхронний повний запит |

## Функції сервера

### Синхронні

| Функція | Опис |
|---------|------|
| `listen(port \| int) \| void` | Почати прослуховування порту |
| `serve(port \| int, handler \| fun) \| void` | Запустити сервер з обробником |

### Асинхронні

| Функція | Опис |
|---------|------|
| `asyncListen(port \| int) \| Future<void>` | Асинхронне прослуховування |

## Хелпери відповідей

| Функція | Опис |
|---------|------|
| `ok(body \| string) \| HttpResponse` | Відповідь 200 OK |
| `notFound() \| HttpResponse` | Відповідь 404 |
| `serverError() \| HttpResponse` | Відповідь 500 |

## Хелпери заголовків і статусів

| Функція | Опис |
|---------|------|
| `setHeader(name \| string, value \| string) \| void` | Встановити заголовок |
| `getHeader(name \| string) \| string` | Отримати заголовок |
| `statusText(code \| int) \| string` | Текст статусу за кодом |

## Приклад клієнта

```avenir
import std.http.client;

fun main() | void {
    var body | string = std.http.client.get("https://example.com");
    print(body);
}
```

### Асинхронний клієнт

```avenir
import std.http.client;

async fun main() | void {
    var body | string = await std.http.client.asyncGet("https://example.com");
    print(body);
}
```

## Обробка помилок

HTTP-операції можуть викидати помилки (мережеві помилки, таймаути тощо):

```avenir
try {
    var body | string = std.http.client.get("https://example.com");
} catch (e | error) {
    print("HTTP помилка: " + e.message());
}
```

## Блокуюча vs неблокуюча поведінка

- Синхронні функції блокують поточний потік до отримання відповіді
- Асинхронні функції виконують I/O у фонових горутинах і повертають future
