# std.crypto.tls — TLS/HTTPS

Модуль `std.crypto.tls` надає підтримку TLS/HTTPS.

## Імпорт

```avenir
import std.crypto.tls;
```

## Структура модуля

```
std/crypto/tls/
├── tls.av           # Головні функції: newConfig, connect, listen
├── config.av        # Config структура
├── certificate.av   # Certificate, PrivateKey
├── connection.av    # Connection (читання/запис)
├── listener.av      # Listener (accept)
└── errors.av        # TLSError
```

## Швидкий старт

### HTTPS-сервер з CoolWeb

```avenir
import std.coolweb;
import std.crypto.tls;

var app | std.coolweb.App = std.coolweb.newApp();

@app.get("/")
async fun home(ctx | std.coolweb.Context) | void {
    ctx.text("Hello over HTTPS!");
}

fun main() | void {
    app.runTLS(443, "cert.pem", "key.pem");
}
```

### HTTPS-клієнт

```avenir
import std.http.client;

async fun main() | void {
    var body | string = await std.http.client.asyncGetTLS("https://example.com");
    print(body);
}
```

### Raw TLS-з'єднання

```avenir
import std.crypto.tls;

async fun main() | void {
    var conn | std.crypto.tls.Connection = await std.crypto.tls.asyncConnect("example.com", 443);
    await conn.asyncWriteString("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n");
    var resp | string = await conn.asyncRead(4096);
    print(resp);
    await conn.asyncClose();
}
```

## Структури

### Config

```avenir
struct Config {
    minVersion     | string     // "1.2" або "1.3"
    maxVersion     | string
    certFile       | string
    keyFile        | string
    alpn           | list<string>
    insecure       | bool       // Пропустити перевірку сертифіката (тільки для тестування!)
    clientAuth     | bool       // Вимагати клієнтський сертифікат
    clientCAFile   | string     // CA для клієнтських сертифікатів
}
```

### Certificate

```avenir
struct Certificate {
    _handle | any
}
```

#### Методи інспекції

| Метод | Опис |
|-------|------|
| `subject() \| string` | Subject сертифіката |
| `issuer() \| string` | Issuer сертифіката |
| `notBefore() \| string` | Початок дії |
| `notAfter() \| string` | Закінчення дії |
| `dnsNames() \| list<string>` | DNS-імена |
| `isCA() \| bool` | Чи є CA-сертифікатом |

### Connection

```avenir
struct Connection {
    _handle | any
}
```

#### Синхронні методи

| Метод | Опис |
|-------|------|
| `read(size \| int) \| string` | Прочитати дані |
| `write(data \| bytes) \| int` | Записати байти |
| `writeString(data \| string) \| int` | Записати рядок |
| `close() \| void` | Закрити з'єднання |
| `peerCertificates() \| list<Certificate>` | Сертифікати peer-а |
| `negotiatedProtocol() \| string` | Узгоджений ALPN-протокол |
| `tlsVersion() \| string` | Версія TLS |

#### Асинхронні методи

| Метод | Опис |
|-------|------|
| `asyncRead(size \| int) \| Future<string>` | Асинхронне читання |
| `asyncWrite(data \| bytes) \| Future<int>` | Асинхронний запис |
| `asyncWriteString(data \| string) \| Future<int>` | Асинхронний запис рядка |
| `asyncClose() \| Future<void>` | Асинхронне закриття |

### Listener

```avenir
struct Listener {
    _handle | any
}
```

| Метод | Опис |
|-------|------|
| `accept() \| Connection` | Прийняти з'єднання |
| `asyncAccept() \| Future<Connection>` | Асинхронне прийняття |
| `close() \| void` | Закрити listener |
| `asyncClose() \| Future<void>` | Асинхронне закриття |

## Функції

### Конфігурація

| Функція | Опис |
|---------|------|
| `newConfig() \| Config` | Створити конфігурацію з безпечними значеннями за замовчуванням |

### Завантаження сертифікатів

| Функція | Опис |
|---------|------|
| `loadCertificate(certFile \| string, keyFile \| string) \| Certificate` | Завантажити з файлів |
| `loadCertificateChain(certFile \| string, keyFile \| string, caFile \| string) \| Certificate` | Завантажити ланцюжок |
| `loadCertificateFromPEM(certPEM \| bytes, keyPEM \| bytes) \| Certificate` | Завантажити з PEM-даних |

### З'єднання

| Функція | Опис |
|---------|------|
| `connect(host \| string, port \| int) \| Connection` | Синхронне TLS-з'єднання |
| `connectConfig(host \| string, port \| int, config \| Config) \| Connection` | З'єднання з конфігурацією |
| `asyncConnect(host \| string, port \| int) \| Future<Connection>` | Асинхронне з'єднання |
| `asyncConnectConfig(host \| string, port \| int, config \| Config) \| Future<Connection>` | Асинхронне з конфігурацією |

### Прослуховування

| Функція | Опис |
|---------|------|
| `listen(port \| int, certFile \| string, keyFile \| string) \| Listener` | Синхронне прослуховування |
| `listenConfig(port \| int, config \| Config) \| Listener` | Прослуховування з конфігурацією |

## Інтеграція з HTTP-сервером

```avenir
import std.http.server;

std.http.server.listenTLS(443, "cert.pem", "key.pem");
std.http.server.listenAutoTLS(443, "example.com");  // Let's Encrypt
```

## Інтеграція з CoolWeb

```avenir
app.runTLS(443, "cert.pem", "key.pem");
app.runTLSConfig(443, config);
app.runAutoTLS(443, "example.com");
```

### mTLS приклад

```avenir
import std.coolweb;
import std.crypto.tls;

var config | std.crypto.tls.Config = std.crypto.tls.newConfig();
config.certFile = "server-cert.pem";
config.keyFile = "server-key.pem";
config.clientAuth = true;
config.clientCAFile = "ca-cert.pem";

app.runTLSConfig(443, config);
```

## Безпека за замовчуванням

- Мінімальна версія TLS: 1.2
- Максимальна версія TLS: 1.3
- ALPN: `["h2", "http/1.1"]`
- Forward secrecy (ECDHE)
- Без небезпечних шифрів

## Продуктивність

- Повторне використання TLS-сесій
- Ефективний handshake через Go `crypto/tls`
- Let's Encrypt автосертифікати з кешуванням
