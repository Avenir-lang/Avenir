# std.http

`std.http` provides a minimal, typed HTTP client and server API built on top of
runtime HTTP primitives.

## Overview

- **Client**: `std.http.client`
- **Server**: `std.http.server`
- **Headers helpers**: `std.http.headers`
- **Status helpers**: `std.http.status`
- **Headers**: `dict<string>`
- **Bodies**: `bytes` (client request body is `bytes?`)

All errors are runtime errors and can be caught with `try / catch`.

## Client API

### Types

```avenir
pub struct HttpRequest {
    method | string
    url | string
    headers | dict<string>
    body | bytes?
}

pub struct HttpResponse {
    status | int
    headers | dict<string>
    body | bytes
}
```

### Functions

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `request` | `req | HttpRequest` | `HttpResponse` | network/protocol errors |
| `get` | `url | string` | `HttpResponse` | network/protocol errors |
| `post` | `url | string`, `body | bytes` | `HttpResponse` | network/protocol errors |
| `put` | `url | string`, `body | bytes` | `HttpResponse` | network/protocol errors |
| `delete` | `url | string` | `HttpResponse` | network/protocol errors |

### Response Helpers

| Method | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `text` | — | `string` | invalid UTF-8 |
| `json` | — | `any` | invalid JSON |

## Server API

### Types

```avenir
pub struct HttpServer {
    handle | any
}

pub struct HttpRequest {
    method | string
    path | string
    headers | dict<string>
    body | bytes
}

pub struct HttpResponse {
    status | int
    headers | dict<string>
    body | bytes
}
```

### Functions

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `listen` | `host | string`, `port | int` | `HttpServer` | bind errors |
| `serve` | `handler | fun(HttpRequest) | HttpResponse` | `void` | accept/handler errors |

### Convenience Responses

| Function | Parameters | Returns | Errors |
| --- | --- | --- | --- |
| `text` | `status | int`, `body | string` | `HttpResponse` | — |
| `json` | `status | int`, `value | any` | `HttpResponse` | JSON stringify errors |

## Headers and Status Helpers

```avenir
import std.http.headers as headers;
import std.http.status as status;

var h | dict<string> = headers.empty();
var ok | int = status.ok();
```

## Utilities

`std.http.utils` contains small helpers:

| Function | Parameters | Returns | Notes |
| --- | --- | --- | --- |
| `ensureHeaders` | `headers | dict<string>` | `dict<string>` | Pass-through helper |

## Error Helpers

Client and server modules include simple error helpers:

| Module | Function | Parameters | Returns |
| --- | --- | --- | --- |
| `std.http.client` | `httpError` | `msg | string` | `error` |
| `std.http.server` | `httpError` | `msg | string` | `error` |

## Error Handling

All parsing, network, and protocol errors surface as runtime errors and can be
caught with `try / catch`.

## Blocking Behavior

`serve()` is blocking and handles one request per accept. This is intentionally
minimal; higher-level concurrency patterns can be built in Avenir.

## Imports

By default, the local alias for an import is the last path segment. To use
`http` as the alias, import with an explicit alias:

```avenir
import std.http.client as http;
import std.http.server as http;
```

## Future Notes

This layout leaves room for middleware, async I/O, HTTPS, and WebSocket support
without changing the core API shape.
