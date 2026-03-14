# std.web.session

Cookie and session management for the Avenir web ecosystem. Framework-agnostic core with built-in `coolweb` integration.

## Quick Start

```avenir
import std.coolweb;
import std.web.session as sess;

var app | coolweb.App = coolweb.newApp();
var memStore | sess.MemoryStore = sess.newMemoryStore();
var adapter | sess.StoreAdapter = memStore.toAdapter();
var config | sess.SessionConfig = sess.newConfigWithOptions(
    "sid", 3600, "my-secret-key", true, true, "Lax", "/", "", "store", 0, adapter
);

app.use(coolweb.sessionMiddleware(config));

@app.get("/")
fun index(ctx | coolweb.Context) | coolweb.Response {
    var s | sess.Session = ctx.session;
    var visits | any = s.get("visits");
    if (visits == none) {
        visits = 0;
    }
    visits = visits + 1;
    s.set("visits", visits);
    return ctx.json({ "visits": visits });
}
```

---

## Architecture

`std.web.session` has **zero dependency on `std.coolweb`**. It is a standalone library.

- `errors.av` — error types
- `config.av` — `SessionConfig` struct + constructors
- `cookies.av` — `Cookie` struct, serialization, parsing
- `session.av` — `Session` struct + API
- `store.av` — `StoreAdapter`, `MemoryStore`, `FileStore`
- `middleware.av` — `SessionManager` (framework-agnostic lifecycle)

Coolweb integration is provided by `std/coolweb/session.av` which imports `std.web.session`.

---

## Cookie API

### Cookie struct

```avenir
pub struct Cookie {
    pub name | string
    pub value | string
    pub path | string = "/"
    pub domain | string = ""
    pub httpOnly | bool = true
    pub secure | bool = true
    pub sameSite | string = "Lax"
    pub maxAge | int = 0
}
```

### Functions

| Function | Description |
|----------|-------------|
| `newCookie(name, value)` | Create a cookie with secure defaults |
| `newCookieWithOptions(name, value, path, domain, httpOnly, secure, sameSite, maxAge)` | Create a cookie with all options |
| `serializeCookie(c)` | Serialize to `Set-Cookie` header string |
| `parseCookies(header)` | Parse a `Cookie:` header into `dict<string>` |
| `deleteCookieString(name, path="/")` | Generate a delete-cookie `Set-Cookie` string |

### Example

```avenir
var c | sess.Cookie = sess.newCookie("theme", "dark");
var header | string = sess.serializeCookie(c);
// theme=dark; Path=/; HttpOnly; Secure; SameSite=Lax
```

---

## Session API

### Session struct

```avenir
pub mut struct Session {
    pub mut id | string
    pub mut data | dict<any>
    pub mut createdAt | int
    pub mut expiresAt | int
    pub mut _changed | bool
    pub mut _destroyed | bool
    pub mut _isNew | bool
}
```

### Methods

| Method | Description |
|--------|-------------|
| `s.get(key)` | Get value by key, returns `none` if missing |
| `s.set(key, value)` | Set value, marks session as changed |
| `s.delete(key)` | Remove key, marks as changed |
| `s.has(key)` | Check if key exists |
| `s.clear()` | Remove all data, marks as changed |
| `s.regenerate()` | Generate new session ID (fixation protection) |
| `s.destroy()` | Mark session for destruction |
| `s.touch(ttlSeconds)` | Extend expiration |
| `s.isNew()` | True if session was just created |
| `s.isChanged()` | True if data was modified |
| `s.isDestroyed()` | True if session was destroyed |
| `s.isExpired()` | True if session has expired |

### Functions

| Function | Description |
|----------|-------------|
| `newSession(ttlSeconds)` | Create a new session with random ID |
| `generateSessionId()` | Generate a 32-byte CSPRNG base64url session ID |

---

## Store API

### StoreAdapter

A type-safe adapter with function-typed fields that any store implementation can provide:

```avenir
pub struct StoreAdapter {
    pub loadFn | fun(string) | any
    pub saveFn | fun(string, dict<any>, int, int) | void
    pub deleteFn | fun(string) | void
}
```

### MemoryStore

In-memory session storage using `dict<any>`. Suitable for development and single-process deployments.

```avenir
var memStore | sess.MemoryStore = sess.newMemoryStore();
var adapter | sess.StoreAdapter = memStore.toAdapter();
```

| Method | Description |
|--------|-------------|
| `load(id)` | Load session entry, returns `none` if expired/missing |
| `save(id, data, createdAt, expiresAt)` | Save session data |
| `deleteSession(id)` | Remove session |
| `cleanup()` | Remove all expired sessions |
| `count()` | Number of stored sessions |
| `toAdapter()` | Convert to `StoreAdapter` |

### FileStore

File-based session storage. One JSON file per session in the specified directory.

```avenir
var fileStore | sess.FileStore = sess.newFileStore("/tmp/sessions");
var adapter | sess.StoreAdapter = fileStore.toAdapter();
```

Same methods as `MemoryStore` (except `cleanup` and `count`).

---

## SessionManager

Framework-agnostic session lifecycle manager. Used internally by the coolweb middleware, but can also be used directly for custom framework integrations.

```avenir
pub mut struct SessionManager {
    pub config | SessionConfig
}
```

| Method | Description |
|--------|-------------|
| `loadSession(cookieValue, adapter)` | Load or create a session |
| `saveSession(session, adapter)` | Save session, returns `Set-Cookie` string |
| `destroySession(session, adapter)` | Destroy session, returns delete-cookie string |
| `buildSessionCookie(value)` | Build a `Set-Cookie` string from config |

---

## SessionConfig

```avenir
pub struct SessionConfig {
    pub cookieName | string = "sid"
    pub ttlSeconds | int = 3600
    pub secret | string = ""
    pub secureCookies | bool = true
    pub httpOnly | bool = true
    pub sameSite | string = "Lax"
    pub path | string = "/"
    pub domain | string = ""
    pub mode | string = "store"          // "store" or "jwt"
    pub idleTimeoutSeconds | int = 0     // 0 = disabled
    pub sessionStore | StoreAdapter
}
```

### Constructors

```avenir
// JWT mode (no store needed)
var config | sess.SessionConfig = sess.newConfig("my-secret");

// Full options
var config | sess.SessionConfig = sess.newConfigWithOptions(
    "sid",          // cookieName
    3600,           // ttlSeconds
    "secret",       // secret
    true,           // secureCookies
    true,           // httpOnly
    "Lax",          // sameSite
    "/",            // path
    "",             // domain
    "store",        // mode
    0,              // idleTimeoutSeconds
    adapter         // sessionStore
);
```

---

## Coolweb Middleware

The `coolweb.sessionMiddleware` function provides seamless integration:

```avenir
app.use(coolweb.sessionMiddleware(config));
```

The middleware:
1. Reads the session cookie from `ctx.cookies`
2. Loads the session via `SessionManager` (store or JWT mode)
3. Attaches the session to `ctx.session`
4. Calls the next handler
5. After the handler: saves if changed, sets `Set-Cookie` header

Access the session in handlers:

```avenir
fun handler(ctx | coolweb.Context) | coolweb.Response {
    var s | sess.Session = ctx.session;
    s.set("userId", 42);
    return ctx.json({ "ok": true });
}
```

---

## JWT Mode

Stateless sessions stored entirely in a signed JWT cookie. No store required.

```avenir
var config | sess.SessionConfig = sess.newConfig("jwt-secret-key");
// config.mode is "store" by default, override via newConfigWithOptions with mode="jwt"
```

JWT payload contains: `sid`, `data`, `iat`, `exp`. Signed with HS256 via `std.crypto.jwt`.

---

## Security

- **Session IDs**: 32 bytes from `crypto/rand`, base64url encoded
- **Cookie defaults**: `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`
- **Session fixation protection**: `s.regenerate()` creates a new ID
- **Expiration**: absolute TTL + optional idle timeout
- **JWT integrity**: HMAC-SHA256 signature verification
- **Cookie tampering**: store mode validates ID against store; JWT mode verifies signature
