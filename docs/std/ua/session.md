# std.web.session — Cookies та сесії

Модуль `std.web.session` надає управління cookies та сесіями.

## Імпорт

```avenir
import std.web.session;
```

## Швидкий старт

```avenir
import std.web.session;
import std.coolweb;

var app | std.coolweb.App = std.coolweb.newApp();

var config | std.web.session.SessionConfig = std.web.session.newConfig("my-secret-key");

@app.use()
async fun sessionMW(ctx | std.coolweb.Context) | void {
    std.coolweb.sessionMiddleware(config);
}

@app.get("/")
async fun home(ctx | std.coolweb.Context) | void {
    ctx.text("Hello!");
}
```

## Архітектура

`std.web.session` має **нульову залежність** від `std.coolweb`. Основні компоненти (Cookie, Session, Store) працюють незалежно. Інтеграція з CoolWeb — опціональна.

## Cookie API

### Структура Cookie

```avenir
struct Cookie {
    name     | string
    value    | string
    path     | string
    domain   | string
    maxAge   | int
    secure   | bool
    httpOnly | bool
    sameSite | string
}
```

### Функції

| Функція | Опис |
|---------|------|
| `newCookie(name \| string, value \| string) \| Cookie` | Створити cookie з безпечними значеннями за замовчуванням |
| `newCookieWithOptions(...) \| Cookie` | Створити cookie з усіма опціями |
| `serializeCookie(c \| Cookie) \| string` | Серіалізувати в рядок Set-Cookie |
| `parseCookies(header \| string) \| list<Cookie>` | Парсинг заголовка Cookie |
| `deleteCookieString(name \| string) \| string` | Рядок для видалення cookie |

### Приклад

```avenir
import std.web.session;

fun main() | void {
    var c | std.web.session.Cookie = std.web.session.newCookie("session_id", "abc123");
    var header | string = std.web.session.serializeCookie(c);
    print(header);
}
```

## Session API

### Структура Session

```avenir
mut struct Session {
    pub mut id        | string
    pub mut data      | dict<any>
    pub mut createdAt | int
    pub mut expiresAt | int
    pub mut isNew     | bool
    // ...
}
```

### Методи Session

| Метод | Опис |
|-------|------|
| `get(key \| string) \| any` | Отримати значення |
| `set(key \| string, value \| any) \| void` | Встановити значення |
| `has(key \| string) \| bool` | Перевірити наявність ключа |
| `deleteSession(key \| string) \| void` | Видалити ключ |
| `clear() \| void` | Очистити всі дані |
| `regenerate() \| void` | Перегенерувати ID сесії |
| `destroy() \| void` | Знищити сесію |
| `touch() \| void` | Оновити час закінчення |
| `isExpired() \| bool` | Перевірити закінчення терміну |
| `isChanged() \| bool` | Перевірити наявність змін |
| `isDestroyed() \| bool` | Перевірити, чи знищена |

### Функції

| Функція | Опис |
|---------|------|
| `newSession(ttl \| int) \| Session` | Створити нову сесію з TTL (секунди) |
| `generateSessionId() \| string` | Згенерувати криптографічно безпечний ID |

## Store API

### StoreAdapter

```avenir
struct StoreAdapter {
    loadFn   | fun(string) | Session?
    saveFn   | fun(Session) | void
    deleteFn | fun(string) | void
}
```

### MemoryStore

Зберігає сесії в пам'яті (для розробки):

```avenir
var store | std.web.session.MemoryStore = std.web.session.MemoryStore{};
var adapter | std.web.session.StoreAdapter = store.toAdapter();
```

### FileStore

Зберігає сесії у файлах:

```avenir
var store | std.web.session.FileStore = std.web.session.FileStore{dir = "./sessions"};
var adapter | std.web.session.StoreAdapter = store.toAdapter();
```

## SessionConfig

```avenir
struct SessionConfig {
    secret    | string
    cookieName | string
    maxAge    | int
    secure    | bool
    httpOnly  | bool
    sameSite  | string
    mode      | string      // "store" або "jwt"
    store     | StoreAdapter
}
```

### Конструктори

| Функція | Опис |
|---------|------|
| `newConfig(secret \| string) \| SessionConfig` | Конфігурація з значеннями за замовчуванням |
| `newConfigWithOptions(...) \| SessionConfig` | Конфігурація з усіма опціями |

## SessionManager

```avenir
struct SessionManager {
    config | SessionConfig
}
```

### Методи

| Метод | Опис |
|-------|------|
| `loadSession(cookieHeader \| string) \| Session` | Завантажити сесію з cookie |
| `saveSession(session \| Session) \| string` | Зберегти і повернути Set-Cookie |
| `destroySession(session \| Session) \| string` | Знищити і повернути Set-Cookie для видалення |

## Інтеграція з CoolWeb

```avenir
import std.coolweb;
import std.web.session;

var config | std.web.session.SessionConfig = std.web.session.newConfig("secret");
std.coolweb.sessionMiddleware(config);
```

Middleware автоматично:
1. Завантажує сесію з cookie запиту
2. Прикріплює до `ctx.session`
3. Зберігає/знищує сесію після обробника

## JWT-режим

При `mode = "jwt"` дані сесії кодуються прямо в cookie як JWT-токен (без серверного сховища):

```avenir
var config | std.web.session.SessionConfig = std.web.session.newConfigWithOptions(
    secret = "jwt-secret",
    mode = "jwt"
);
```

## Безпека

- ID сесій: 32 байти CSPRNG, base64url
- Cookie за замовчуванням: `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`
- JWT підписується HMAC-SHA256
- Автоматичне закінчення терміну сесій
