# std.coolweb — Асинхронний веб-фреймворк

`std.coolweb` — асинхронний веб-фреймворк для Avenir.

## Імпорт

```avenir
import std.coolweb;
```

## Швидкий старт

```avenir
import std.coolweb;

var app | std.coolweb.App = std.coolweb.newApp();

@app.get("/")
async fun home(ctx | std.coolweb.Context) | void {
    ctx.text("Hello, World!");
}

@app.get("/users/:id")
async fun getUser(ctx | std.coolweb.Context) | void {
    var id | string = ctx.param("id");
    ctx.json("{\"id\": \"" + id + "\"}");
}

fun main() | void {
    app.run(8080);
}
```

## App

### Створення

```avenir
var app | std.coolweb.App = std.coolweb.newApp();
```

### Реєстрація маршрутів через декоратори

Маршрути реєструються через методи-декоратори на `App`:

```avenir
@app.get("/path")
async fun handler(ctx | std.coolweb.Context) | void { ... }

@app.post("/path")
async fun handler(ctx | std.coolweb.Context) | void { ... }

@app.put("/path")
async fun handler(ctx | std.coolweb.Context) | void { ... }

@app.delete("/path")
async fun handler(ctx | std.coolweb.Context) | void { ... }
```

### Middleware

```avenir
@app.use()
async fun logger(ctx | std.coolweb.Context) | void {
    print("Запит: " + ctx.method() + " " + ctx.path());
    ctx.next();
}
```

Middleware виконується в порядку реєстрації. Виклик `ctx.next()` передає керування наступному middleware або обробнику.

### Обробка помилок

```avenir
@app.onError()
async fun errorHandler(ctx | std.coolweb.Context) | void {
    ctx.status(500);
    ctx.text("Внутрішня помилка сервера");
}
```

### Запуск сервера

```avenir
app.run(8080);
```

Для HTTPS:

```avenir
app.runTLS(443, "cert.pem", "key.pem");
app.runAutoTLS(443, "example.com");
```

## Context

`Context` надає доступ до даних запиту та побудови відповіді.

### Дані запиту

| Метод | Опис |
|-------|------|
| `method() \| string` | HTTP-метод |
| `path() \| string` | Шлях запиту |
| `param(name \| string) \| string` | Параметр шляху |
| `query(name \| string) \| string` | Query-параметр |
| `header(name \| string) \| string` | Заголовок запиту |
| `body() \| string` | Тіло запиту |

### Побудова відповіді

| Метод | Опис |
|-------|------|
| `text(body \| string) \| void` | Текстова відповідь |
| `json(body \| string) \| void` | JSON-відповідь |
| `html(body \| string) \| void` | HTML-відповідь |
| `status(code \| int) \| void` | Встановити статус-код |
| `setHeader(name \| string, value \| string) \| void` | Встановити заголовок відповіді |
| `redirect(url \| string) \| void` | Перенаправлення |

### Парсери тіла

| Метод | Опис |
|-------|------|
| `jsonBody() \| dict<any>` | Парсинг JSON-тіла |
| `formBody() \| dict<string>` | Парсинг form-тіла |

### Керування потоком

| Метод | Опис |
|-------|------|
| `next() \| void` | Передати наступному middleware |

## Router

Вкладені маршрути підтримуються через Router:

```avenir
var api | std.coolweb.Router = std.coolweb.newRouter("/api");

@api.get("/users")
async fun listUsers(ctx | std.coolweb.Context) | void {
    ctx.json("[]");
}

app.mount(api);
```

## Параметри шляху

Динамічні сегменти шляху позначаються `:`:

```avenir
@app.get("/users/:id")
async fun getUser(ctx | std.coolweb.Context) | void {
    var id | string = ctx.param("id");
    ctx.text("Користувач: " + id);
}
```

## Вбудований middleware

### Logger

```avenir
import std.coolweb;

@app.use()
async fun logger(ctx | std.coolweb.Context) | void {
    print(ctx.method() + " " + ctx.path());
    ctx.next();
}
```

### CORS

```avenir
import std.coolweb;

var corsConfig | dict<string> = {
    origin: "*",
    methods: "GET,POST,PUT,DELETE",
    headers: "Content-Type,Authorization"
};

app.cors(corsConfig);
```

## Конструктори відповідей

| Функція | Опис |
|---------|------|
| `textResponse(body \| string, status \| int) \| void` | Текстова відповідь зі статусом |
| `jsonResponse(body \| string, status \| int) \| void` | JSON-відповідь зі статусом |
| `htmlResponse(body \| string, status \| int) \| void` | HTML-відповідь зі статусом |

## Структура модуля

```
std/coolweb/
├── coolweb.av       # App, запуск сервера, диспетчеризація
├── context.av       # Context — запит/відповідь
├── router.av        # Router, зіставлення маршрутів
├── cors.av          # CORS middleware
├── response.av      # Хелпери відповідей
├── session.av       # Інтеграція сесій
└── ws_route.av      # WebSocket маршрути
```
