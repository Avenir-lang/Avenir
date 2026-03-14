# std.net — TCP мережа

Модуль `std.net` надає TCP-з'єднання та серверне API.

## Імпорт

```avenir
import std.net;
```

## Структури

### Socket

```avenir
struct Socket {
    _handle | any
}
```

### Server

```avenir
struct Server {
    _handle | any
}
```

## Функції

### Синхронні

| Функція | Опис |
|---------|------|
| `connect(host \| string, port \| int) \| Socket` | Підключитися до TCP-сервера |
| `listen(host \| string, port \| int) \| Server` | Почати прослуховування TCP |

### Асинхронні

| Функція | Опис |
|---------|------|
| `asyncConnect(host \| string, port \| int) \| Future<Socket>` | Асинхронне підключення |

## Методи Socket

### Синхронні

| Метод | Опис |
|-------|------|
| `read(size \| int) \| string` | Прочитати до `size` байтів |
| `write(data \| string) \| void` | Записати дані |
| `close() \| void` | Закрити з'єднання |

### Асинхронні

| Метод | Опис |
|-------|------|
| `asyncRead(size \| int) \| Future<string>` | Асинхронне читання |
| `asyncWrite(data \| string) \| Future<void>` | Асинхронний запис |
| `asyncClose() \| Future<void>` | Асинхронне закриття |

## Методи Server

### Синхронні

| Метод | Опис |
|-------|------|
| `accept() \| Socket` | Прийняти вхідне з'єднання |
| `close() \| void` | Зупинити сервер |

### Асинхронні

| Метод | Опис |
|-------|------|
| `asyncAccept() \| Future<Socket>` | Асинхронне прийняття з'єднання |

## Приклад

### TCP клієнт

```avenir
import std.net;

fun main() | void {
    var sock | std.net.Socket = std.net.connect("127.0.0.1", 8080);
    sock.write("Hello, Server!");
    var response | string = sock.read(1024);
    print(response);
    sock.close();
}
```

### TCP сервер

```avenir
import std.net;

fun main() | void {
    var server | std.net.Server = std.net.listen("0.0.0.0", 8080);
    print("Сервер слухає на порту 8080");

    var client | std.net.Socket = server.accept();
    var data | string = client.read(1024);
    print("Отримано: " + data);
    client.write("Echo: " + data);
    client.close();
    server.close();
}
```

### Асинхронний приклад

```avenir
import std.net;

async fun main() | void {
    var sock | std.net.Socket = await std.net.asyncConnect("127.0.0.1", 8080);
    await sock.asyncWrite("Hello!");
    var resp | string = await sock.asyncRead(1024);
    print(resp);
    await sock.asyncClose();
}
```

## Обробка помилок

Мережеві операції можуть викидати помилки (з'єднання відхилено, таймаут тощо):

```avenir
try {
    var sock | std.net.Socket = std.net.connect("127.0.0.1", 9999);
} catch (e | error) {
    print("Помилка з'єднання: " + e.message());
}
```

## Блокуюча vs неблокуюча поведінка

- Синхронні функції (`connect`, `read`, `write`) блокують поточний потік
- Асинхронні функції (`asyncConnect`, `asyncRead`, `asyncWrite`) виконують I/O у фонових горутинах і повертають future
