# std.crypto.tls — TLS/HTTPS Support

The `std.crypto.tls` module provides TLS (Transport Layer Security) primitives for secure network communication. It integrates with `std.http.server`, `std.http.client`, and `std.coolweb`.

## Module Structure

```
std/crypto/tls/
  tls.av          — Core functions: newConfig, loadCertificate, connect, listen
  config.av       — Config struct
  certificate.av  — Certificate, PrivateKey structs + inspection methods
  connection.av   — TLS Connection (read/write/close)
  listener.av     — TLS Listener (accept/close)
  errors.av       — TLSError struct
```

## Quick Start

### HTTPS Server with CoolWeb

```avenir
import std.coolweb;

async fun main() | void {
    var app | coolweb.App = coolweb.newApp();

    app.get("/", fun(ctx | coolweb.Context) | coolweb.Response {
        return ctx.text("Hello, HTTPS!");
    });

    await app.runTLS(443, "cert.pem", "key.pem");
}
```

### HTTPS Client

```avenir
import std.http.client as http;

async fun main() | void {
    var resp | http.HttpResponse = await http.asyncGet("https://example.com");
    print("Status: ${resp.status}");
    print("Body: ${resp.text()}");
}
```

### Raw TLS Connection

```avenir
import std.crypto.tls;

async fun main() | void {
    var conn | tls.Connection = await tls.asyncConnect("example.com", 443);
    print("TLS: ${conn.tlsVersion()}");
    await conn.asyncWriteString("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n");
    var data | bytes = await conn.asyncRead(4096);
    print(data.toString());
    await conn.asyncClose();
}
```

---

## Structs

### Config

TLS configuration for servers and clients.

```avenir
pub struct Config {
    certFile | string           — Path to PEM certificate file
    keyFile | string            — Path to PEM private key file
    minVersion | string         — Minimum TLS version ("1.2", "1.3")
    maxVersion | string         — Maximum TLS version ("1.2", "1.3")
    alpnProtocols | list<string> — ALPN protocol list (default: ["h2", "http/1.1"])
    clientAuth | string         — Client auth mode ("none", "request", "require", "verify", "requireAndVerify")
    clientCAs | list<string>    — Paths to client CA certificate files
    insecureSkipVerify | bool   — Skip server certificate verification (testing only!)
    serverName | string         — Override server name for verification
}
```

### Certificate

Loaded X.509 certificate with inspection methods.

```avenir
pub struct Certificate { handle | any }

cert.subject() | string
cert.issuer() | string
cert.notBefore() | int          — Unix timestamp
cert.notAfter() | int           — Unix timestamp
cert.dnsNames() | list<string>
cert.isCA() | bool
```

### Connection

TLS-encrypted network connection.

```avenir
pub struct Connection { handle | any }

conn.read(n | int) | bytes
conn.write(data | bytes) | int
conn.writeString(data | string) | int
conn.close() | void
conn.peerCertificates() | list<dict<string>>
conn.negotiatedProtocol() | string
conn.tlsVersion() | string

conn.asyncRead(n | int) | bytes
conn.asyncWrite(data | bytes) | int
conn.asyncWriteString(data | string) | int
conn.asyncClose() | void
```

### Listener

TLS server listener.

```avenir
pub struct Listener { handle | any }

ln.accept() | Connection
ln.close() | void
ln.asyncAccept() | Connection
ln.asyncClose() | void
```

---

## Functions

### Configuration

```avenir
tls.newConfig(certFile | string, keyFile | string) | Config
```

Creates a Config with secure defaults (TLS 1.2+, ALPN h2/http1.1).

### Certificate Loading

```avenir
tls.loadCertificate(certFile | string, keyFile | string) | Certificate
tls.loadCertificateChain(files | list<string>) | Certificate
tls.loadCertificateFromPEM(certPEM | bytes, keyPEM | bytes) | Certificate
```

### Connect (Client)

```avenir
tls.connect(host | string, port | int) | Connection
tls.connectConfig(host | string, port | int, cfg | Config) | Connection
tls.asyncConnect(host | string, port | int) | Connection
tls.asyncConnectConfig(host | string, port | int, cfg | Config) | Connection
```

### Listen (Server)

```avenir
tls.listen(host | string, port | int, certFile | string, keyFile | string) | Listener
tls.listenConfig(host | string, port | int, cfg | Config) | Listener
```

---

## HTTP Server Integration

`std.http.server` gains three new functions:

```avenir
http.listenTLS(host, port, certFile, keyFile) | HttpServer
http.listenTLSConfig(host, port, cfg) | HttpServer
http.listenAutoTLS(host, port, domain, email) | HttpServer
```

After calling `listenTLS`, `asyncAccept()` and `rawRespond()` work identically to plain HTTP — the TLS layer is transparent.

---

## CoolWeb Integration

`std.coolweb.App` gains three new methods:

```avenir
app.runTLS(port, certFile, keyFile)         — HTTPS with certificate files
app.runTLSConfig(port, cfg)                 — HTTPS with Config dict
app.runAutoTLS(port, domain, email)         — Automatic Let's Encrypt
```

### Example: Auto-TLS Production Server

```avenir
import std.coolweb;

async fun main() | void {
    var app | coolweb.App = coolweb.newApp();
    app.get("/", fun(ctx | coolweb.Context) | coolweb.Response {
        return ctx.text("Secured by Let's Encrypt!");
    });
    await app.runAutoTLS(443, "example.com", "admin@example.com");
}
```

### Example: Mutual TLS (mTLS)

```avenir
import std.crypto.tls;
import std.coolweb;

async fun main() | void {
    var cfg | tls.Config = tls.newConfig("server.pem", "server-key.pem");
    cfg.clientAuth = "requireAndVerify";
    cfg.clientCAs = ["client-ca.pem"];

    var app | coolweb.App = coolweb.newApp();
    app.get("/secure", fun(ctx | coolweb.Context) | coolweb.Response {
        return ctx.json({"authenticated": true});
    });
    await app.runTLSConfig(443, cfg.toDict());
}
```

---

## HTTP Client HTTPS

`std.http.client` HTTPS requests work transparently:

```avenir
var resp | http.HttpResponse = await http.asyncGet("https://api.example.com/data");
```

For custom TLS configuration (e.g., self-signed certs, client certificates):

```avenir
http.asyncRequestTLS(req, cfg) | HttpResponse
http.asyncGetTLS(url, cfg) | HttpResponse
http.asyncPostTLS(url, body, cfg) | HttpResponse
```

The `cfg` parameter is a `dict<any>` — use `tls.newConfig(...).toDict()` to create it.

---

## Security Defaults

| Setting | Default |
|---|---|
| Min TLS version | 1.2 |
| Max TLS version | 1.3 |
| Cipher suites | Go defaults (AES-GCM, ChaCha20-Poly1305) |
| ALPN | ["h2", "http/1.1"] |
| Forward secrecy | Enforced (ECDHE only) |
| Client auth | None (opt-in) |
| Certificate verification | Enabled (system root CAs) |

---

## Performance Notes

- All I/O operations have async variants that don't block the event loop
- TLS session tickets enabled by default for resumption
- HTTP client connection pooling works transparently for HTTPS
- Go's `crypto/tls` handles buffer pooling internally
