# std.websocket — WebSocket

Модуль `std.websocket` надає підтримку протоколу WebSocket.

## Імпорт

```avenir
import std.websocket;
```

## Константи

Типи повідомлень доступні через функції-геттери (в Avenir немає `pub var`):

| Функція | Опис |
|---------|------|
| `msgText() \| int` | Текстове повідомлення |
| `msgBinary() \| int` | Бінарне повідомлення |
| `msgClose() \| int` | Повідомлення закриття |
| `msgPing() \| int` | Ping |
| `msgPong() \| int` | Pong |

## Структури

### WebSocket

```avenir
pub mut struct WebSocket {
    pub id         | string
    pub path       | string
    pub headers    | dict<string>
    pub query      | dict<string>
    pub remoteAddr | string
    pub mut isOpen | bool
    _handle        | any
}
```

### Message

```avenir
pub struct Message {
    pub msgType | int
    pub data    | bytes
}
```

#### Методи Message

| Метод | Опис |
|-------|------|
| `isText() \| bool` | Чи текстове |
| `isBinary() \| bool` | Чи бінарне |
| `isClose() \| bool` | Чи закриття |
| `isPing() \| bool` | Чи ping |
| `isPong() \| bool` | Чи pong |

### Hub

```avenir
pub mut struct Hub {
    // внутрішні поля
}
```

## Функції

| Функція | Опис |
|---------|------|
| `upgradeRequest(reqHandle \| any, protocols \| list<string>, headers \| dict<string>) \| Future<WebSocket>` | Оновити HTTP-запит до WebSocket |
| `newHub() \| Hub` | Створити новий Hub |
| `newMessage(msgType \| int, data \| bytes) \| Message` | Створити повідомлення |

## Методи WebSocket

| Метод | Опис |
|-------|------|
| `sendText(data \| string) \| Future<void>` | Надіслати текст |
| `sendBytes(data \| bytes) \| Future<void>` | Надіслати байти |
| `sendPing(data \| bytes) \| Future<void>` | Надіслати ping |
| `receive() \| Future<Message>` | Отримати повідомлення |
| `close() \| Future<void>` | Закрити з'єднання |
| `setReadLimit(limit \| int) \| void` | Встановити максимальний розмір повідомлення |

## Методи Hub

| Метод | Опис |
|-------|------|
| `add(ws \| WebSocket) \| void` | Додати з'єднання |
| `remove(ws \| WebSocket) \| void` | Видалити з'єднання |
| `broadcast(data \| string) \| Future<void>` | Розіслати текст усім |
| `broadcastBytes(data \| bytes) \| Future<void>` | Розіслати байти усім |
| `count() \| int` | Кількість з'єднань |
| `cleanup() \| void` | Видалити закриті з'єднання |

## Типи помилок

| Структура | Опис |
|-----------|------|
| `WebSocketError` | Загальна помилка WebSocket |
| `WebSocketClosed` | З'єднання закрите |
| `ProtocolError` | Помилка протоколу |

## Інтеграція з CoolWeb

### Реєстрація маршрутів

```avenir
import std.coolweb;
import std.websocket as ws;

var app | std.coolweb.App = std.coolweb.newApp();

@app.websocket("/ws/echo")
async fun echoHandler(ctx | std.coolweb.Context, conn | ws.WebSocket) | void {
    while (conn.isOpen) {
        var msg | ws.Message = await conn.receive();
        if (msg.isText()) {
            await conn.sendText(msg.data.toString());
        }
    }
}
```

### Router-рівень

```avenir
var api | std.coolweb.Router = std.coolweb.newRouter("/api");

@api.websocket("/chat")
async fun chatHandler(ctx | std.coolweb.Context, conn | ws.WebSocket) | void {
    // ...
}

app.mount(api);
```

## Приклади

### Echo-сервер

```avenir
import std.coolweb;
import std.websocket as ws;

var app | std.coolweb.App = std.coolweb.newApp();

@app.websocket("/ws")
async fun echo(ctx | std.coolweb.Context, conn | ws.WebSocket) | void {
    while (conn.isOpen) {
        try {
            var msg | ws.Message = await conn.receive();
            if (msg.isText()) {
                await conn.sendText(msg.data.toString());
            }
        } catch (e | ws.WebSocketClosed) {
            print("З'єднання закрито");
        }
    }
}

fun main() | void {
    app.run(8080);
}
```

### Чат з Hub

```avenir
import std.coolweb;
import std.websocket as ws;

var app | std.coolweb.App = std.coolweb.newApp();
var hub | ws.Hub = ws.newHub();

@app.websocket("/ws/chat")
async fun chat(ctx | std.coolweb.Context, conn | ws.WebSocket) | void {
    hub.add(conn);
    while (conn.isOpen) {
        try {
            var msg | ws.Message = await conn.receive();
            if (msg.isText()) {
                await hub.broadcast(msg.data.toString());
            }
        } catch (e | ws.WebSocketClosed) {
            hub.remove(conn);
        }
    }
}
```

## Деталі протоколу

- RFC 6455 фреймовий парсер
- Автоматичний pong на ping
- Коректний handshake закриття
- М'ютекс запису на кожне з'єднання
- Атомарний прапор закриття
- Максимальний розмір повідомлення за замовчуванням: 1 МБ
- Дедлайн запису: 30 секунд
